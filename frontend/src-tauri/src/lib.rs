#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
  tauri::Builder::default()
    .setup(|app| {
      app.handle().plugin(
        tauri_plugin_log::Builder::default()
          .level(log::LevelFilter::Info)
          .build(),
      )?;
      
      // Launch backend API sidecar
      use tauri_plugin_shell::ShellExt;
      let sidecar_command = app.shell().sidecar("backend-api").unwrap();
      let (mut rx, mut _child) = sidecar_command
        .spawn()
        .expect("Failed to spawn backend API");

      use tauri_plugin_shell::process::CommandEvent;
      tauri::async_runtime::spawn(async move {
        while let Some(event) = rx.recv().await {
          if let CommandEvent::Stdout(line) = event {
            log::info!("Sidecar: {}", String::from_utf8_lossy(&line));
          } else if let CommandEvent::Stderr(line) = event {
            log::error!("Sidecar Error: {}", String::from_utf8_lossy(&line));
          }
        }
      });
        
      Ok(())
    })
    .plugin(tauri_plugin_shell::init())
    .run(tauri::generate_context!())
    .expect("error while running tauri application");
}
