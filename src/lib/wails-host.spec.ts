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
  const args: Record<string, unknown[]> = {};

  const runtime: WailsRuntimeLike = {
    Call: {
      async ByName(name: string, ...rest: unknown[]) {
        calls.push(name);
        args[name] = rest;
        if (name === `${DEFAULT_SERVICE}.SetApplicationMenu` || name === `${DEFAULT_SERVICE}.ShowContextMenu`) return undefined;
        if (name === `${DEFAULT_SERVICE}.Platform`) return platform;
        if (name === `${DEFAULT_SERVICE}.Appearance`) return appearance;
        if (name === `${DEFAULT_SERVICE}.OpenFile`) return null; // Go's nil slice: a cancelled dialog
        if (name === `${DEFAULT_SERVICE}.SaveFile`) return { name: 'Report.pdf', path: '/home/me/Report.pdf' };
        if (name === `${DEFAULT_SERVICE}.PickFolder`) return null;
        if (name === `${DEFAULT_SERVICE}.Message`) return 'save';
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
  return { runtime, driver, calls, args, emit };
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

  it('shows native menus through the Go service and reports chosen items', async () => {
    const { runtime, args, emit } = fakeWails();
    const host = wailsHost({ runtime });
    const menu = [{ label: 'File', submenu: [{ id: 'new', label: 'New', accelerator: 'mod+n' }] }];

    await host.menus!.setApplicationMenu!(menu);
    expect(args[`${DEFAULT_SERVICE}.SetApplicationMenu`]).toEqual([menu]);
    await host.menus!.showContextMenu!(menu, { x: 10.6, y: 20.2 });
    expect(args[`${DEFAULT_SERVICE}.ShowContextMenu`]).toEqual([menu, 11, 20]);

    const chosen: string[] = [];
    const off = host.menus!.onAction((id) => chosen.push(id));
    await tick();
    emit('spartan-native:menu-action', 'new');
    off();
    emit('spartan-native:menu-action', 'ignored');
    expect(chosen).toEqual(['new']);
  });

  it('shows the OS file dialogs and alerts through the Go service', async () => {
    const { runtime, args } = fakeWails();
    const host = wailsHost({ runtime });
    const filters = [{ name: 'PDF', extensions: ['pdf'] }];

    expect(await host.dialogs!.openFile({ filters, multiple: true })).toEqual([]);
    expect(args[`${DEFAULT_SERVICE}.OpenFile`]).toEqual([{ filters, multiple: true }]);
    expect(await host.dialogs!.saveFile({ fileName: 'Report.pdf' })).toEqual({ name: 'Report.pdf', path: '/home/me/Report.pdf' });
    expect(await host.dialogs!.pickFolder()).toBeNull();
    expect(args[`${DEFAULT_SERVICE}.PickFolder`]).toEqual([{}]);

    const buttons = [{ id: 'cancel', label: 'Cancel' }, { id: 'save', label: 'Save' }];
    expect(await host.dialogs!.message!({ kind: 'question', message: 'Save changes?', buttons, defaultId: 'save', cancelId: 'cancel' })).toBe('save');
  });
});
