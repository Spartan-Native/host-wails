import {
  type AppearanceState,
  CAPABILITIES,
  type HostCapabilities,
  type NativeHost,
  type PlatformInfo,
  type Unsubscribe,
} from '@spartan-native/theme';

/** The Go service's bound-method prefix: `<Go package path>.<type>`. */
export const DEFAULT_SERVICE = 'github.com/spartan-native/host-wails/native.Service';
/** The event the Go service emits on every appearance change. */
export const APPEARANCE_EVENT = 'spartan-native:appearance';
/** The event the Go service emits with an item's id when a native menu item is chosen. */
export const MENU_ACTION_EVENT = 'spartan-native:menu-action';

/** The parts of @wailsio/runtime this adapter uses. Inject a fake in tests. */
export interface WailsRuntimeLike {
  Call: { ByName(methodName: string, ...args: unknown[]): Promise<unknown> };
  Events: { On(eventName: string, callback: (event: { data: unknown }) => void): () => void };
  Window: {
    Minimise(): Promise<void>;
    ToggleMaximise(): Promise<void>;
    Close(): Promise<void>;
    IsMaximised(): Promise<boolean>;
  };
}

export interface WailsHostOptions {
  /** The Go service's bound name, if you wrapped or renamed it. Default: DEFAULT_SERVICE. */
  service?: string;
  /** Use this runtime instead of importing @wailsio/runtime (tests, custom builds). */
  runtime?: WailsRuntimeLike;
}

/**
 * True when the page runs inside a Wails webview: `wails://` on macOS and Linux,
 * `wails.localhost` on Windows (also in `wails3 dev`).
 */
export function isWails(location: Pick<Location, 'protocol' | 'hostname'> | undefined = globalThis.location): boolean {
  return !!location && (location.protocol === 'wails:' || location.hostname === 'wails.localhost');
}

/** Capabilities until the Go service reports the real ones (the Wails column of the matrix). */
const INITIAL: Partial<HostCapabilities> = {
  systemAccent: 'native', colorScheme: 'native', highContrast: 'native', reduceTransparency: 'native',
  textScale: 'native', windowFocus: 'native', windowMaterial: 'native', customTitleBar: 'native',
  nativeMenus: 'partial', notifications: 'partial', fileDialogs: 'native', osVersion: 'native',
};

/**
 * NativeHost for Go + Wails v3. Reads the real OS settings through the Go service
 * (github.com/spartan-native/host-wails/native), live, and owns the window:
 * desktop-ui's window controls minimize, maximize and close it. Native menus too:
 * desktop-ui's menubar becomes the macOS global menu bar, and its context menus are
 * the OS's own.
 *
 * ```go
 * // main.go
 * Services: []application.Service{application.NewService(spartannative.New())}
 * ```
 * ```ts
 * // app.config.ts
 * provideSpartanNative({ host: wailsHost() })
 * ```
 */
export function wailsHost(options: WailsHostOptions = {}): NativeHost {
  const service = options.service ?? DEFAULT_SERVICE;
  // Loaded once, right away (lazily imported so a plain browser build never runs Wails code).
  let loaded: WailsRuntimeLike | null = options.runtime ?? null;
  const runtimePromise: Promise<WailsRuntimeLike> = loaded
    ? Promise.resolve(loaded)
    : (import('@wailsio/runtime') as Promise<unknown> as Promise<WailsRuntimeLike>).then((rt) => (loaded = rt));
  const runtime = () => runtimePromise;

  const call = async <T>(method: string, ...args: unknown[]): Promise<T> =>
    (await runtime()).Call.ByName(`${service}.${method}`, ...args) as Promise<T>;

  // Starts with the matrix values; replaced in place once the OS reports its own.
  const capabilities = Object.fromEntries(CAPABILITIES.map((c) => [c, INITIAL[c] ?? 'none'])) as HostCapabilities;
  void call<Partial<HostCapabilities>>('Capabilities')
    .then((caps) => Object.assign(capabilities, caps))
    .catch(() => undefined);

  /** Subscribe to a Wails event: at once when the runtime is loaded, otherwise as soon as it is. */
  const on = (event: string, cb: (data: unknown) => void): Unsubscribe => {
    let off: (() => void) | null = null;
    let cancelled = false;
    if (loaded) off = loaded.Events.On(event, (e) => cb(e.data));
    else
      void runtime().then((rt) => {
        if (!cancelled) off = rt.Events.On(event, (e) => cb(e.data));
      });
    return () => {
      cancelled = true;
      off?.();
      off = null;
    };
  };

  return {
    runtime: 'wails',
    capabilities,
    platform: () => call<PlatformInfo>('Platform'),
    appearance: () => call<AppearanceState>('Appearance'),
    onAppearanceChange: (cb) => on(APPEARANCE_EVENT, (data) => cb(data as AppearanceState)),
    window: {
      isFocused: async () => (typeof document === 'undefined' ? true : document.hasFocus()),
      onFocusChange(cb): Unsubscribe {
        const offs = [on('common:WindowFocus', () => cb(true)), on('common:WindowLostFocus', () => cb(false))];
        return () => offs.forEach((off) => off());
      },
      minimize: async () => (await runtime()).Window.Minimise(),
      toggleMaximize: async () => (await runtime()).Window.ToggleMaximise(),
      close: async () => (await runtime()).Window.Close(),
      isMaximized: async () => (await runtime()).Window.IsMaximised(),
      onMaximizedChange(cb): Unsubscribe {
        const offs = [on('common:WindowMaximise', () => cb(true)), on('common:WindowUnMaximise', () => cb(false))];
        return () => offs.forEach((off) => off());
      },
    },
    menus: {
      setApplicationMenu: (menu) => call<void>('SetApplicationMenu', menu),
      showContextMenu: (menu, { x, y }) => call<void>('ShowContextMenu', menu, Math.round(x), Math.round(y)),
      onAction: (cb) => on(MENU_ACTION_EVENT, (id) => cb(String(id))),
    },
  };
}
