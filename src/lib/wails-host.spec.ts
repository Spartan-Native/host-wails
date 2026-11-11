import { describeHostConformance, type HostDriver } from '@spartan-native/host-conformance';
import type { AppearanceState, PlatformInfo } from '@spartan-native/theme';
import { DEFAULT_SERVICE, isWails, wailsHost, type WailsRuntimeLike } from './wails-host';

/**
 * A fake Wails runtime that behaves like the Go service: bound methods answer by
 * fully qualified name, and the driver emits the same events the service would.
 */
function fakeWails(initial: Partial<AppearanceState> = {}) {
  let appearance: AppearanceState = {
    skin: 'win', accent: '#0078d4', scheme: 'light', contrast: 'normal',
    reduceTransparency: false, reduceMotion: false, textScale: 1, ...initial,
  };
  const platform: PlatformInfo = { os: 'windows', version: '10.0.26100', build: 26100, desktop: null, skin: 'win', confidence: 'exact' };
  const listeners = new Map<string, Set<(e: { data: unknown }) => void>>();
  const emit = (name: string, data?: unknown) => listeners.get(name)?.forEach((cb) => cb({ data }));
  const calls: string[] = [];

  const runtime: WailsRuntimeLike = {
    Call: {
      async ByName(name: string) {
        calls.push(name);
        if (name === `${DEFAULT_SERVICE}.Platform`) return platform;
        if (name === `${DEFAULT_SERVICE}.Appearance`) return appearance;
        if (name === `${DEFAULT_SERVICE}.Capabilities`) return { textScale: 'native', taskbarProgress: 'none' };
        throw new Error(`unknown method ${name}`);
      },
    },
    Events: {
      On(name, cb) {
        if (!listeners.has(name)) listeners.set(name, new Set());
        listeners.get(name)!.add(cb);
        return () => listeners.get(name)!.delete(cb);
      },
    },
    Window: {
      Minimise: async () => void calls.push('Minimise'),
      ToggleMaximise: async () => void calls.push('ToggleMaximise'),
      Close: async () => void calls.push('Close'),
      IsMaximised: async () => false,
    },
  };

  const driver: HostDriver = {
    setAppearance(patch) {
      appearance = { ...appearance, ...patch };
      emit('spartan-native:appearance', appearance); // what the Go service does on a change
    },
    setFocused(focused) {
      emit(focused ? 'common:WindowFocus' : 'common:WindowLostFocus');
    },
  };
  return { runtime, driver, calls, emit };
}

const tick = () => new Promise((r) => setTimeout(r));

describeHostConformance('wailsHost', () => {
  const { runtime, driver } = fakeWails();
  return { host: wailsHost({ runtime }), driver };
});

describe('wailsHost', () => {
  it('reads the OS through the Go service by its bound name', async () => {
    const { runtime, calls } = fakeWails({ accent: '#e66100', scheme: 'dark' });
    const host = wailsHost({ runtime });
    expect(await host.appearance()).toMatchObject({ accent: '#e66100', scheme: 'dark' });
    expect((await host.platform()).build).toBe(26100);
    expect(calls).toContain('github.com/spartan-native/host-wails/native.Service.Appearance');
  });

  it('takes the capabilities the OS reports', async () => {
    const { runtime } = fakeWails();
    const host = wailsHost({ runtime });
    await tick();
    expect(host.capabilities.taskbarProgress).toBe('none');
    expect(host.capabilities.systemAccent).toBe('native');
  });

  it('owns the window: minimize, maximize, close and the maximized state', async () => {
    const { runtime, calls, emit } = fakeWails();
    const host = wailsHost({ runtime });
    await host.window.minimize!();
    await host.window.toggleMaximize!();
    await host.window.close!();
    expect(calls.slice(-3)).toEqual(['Minimise', 'ToggleMaximise', 'Close']);

    const states: boolean[] = [];
    const off = host.window.onMaximizedChange!((m) => states.push(m));
    await tick();
    emit('common:WindowMaximise');
    emit('common:WindowUnMaximise');
    off();
    expect(states).toEqual([true, false]);
  });

  it('detects the Wails webview by origin', () => {
    expect(isWails({ protocol: 'wails:', hostname: 'localhost' })).toBe(true);
    expect(isWails({ protocol: 'http:', hostname: 'wails.localhost' })).toBe(true);
    expect(isWails({ protocol: 'http:', hostname: 'localhost' })).toBe(false);
  });
});
