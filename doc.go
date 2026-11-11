// Package spartannative is the native side of @spartan-native/host-wails.
//
// Register the service with your Wails v3 application:
//
//	app := application.New(application.Options{
//		Services: []application.Service{
//			application.NewService(spartannative.New()),
//		},
//	})
//
// The frontend calls Platform(), Appearance() and Capabilities() through the
// Wails bindings, and receives the "spartan-native:appearance" event whenever
// the OS appearance changes (accent, light/dark, contrast, transparency,
// motion, text size). SetApplicationMenu() and ShowContextMenu() show native
// menus; a chosen item comes back as the "spartan-native:menu-action" event.
// OpenFile(), SaveFile(), PickFolder() and Message() show the OS's file dialogs
// and alerts.
package spartannative
