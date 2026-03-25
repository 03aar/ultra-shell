mod context;
mod ipc;
mod pty;
mod rollback;

use crate::pty::interceptor::NucleusShell;
use std::path::PathBuf;

fn main() {
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or("info"))
        .format_timestamp_millis()
        .init();

    log::info!("Starting NUCLEUS — AI-native shell runtime");

    let args: Vec<String> = std::env::args().collect();

    // Determine data directory
    let data_dir = std::env::var("NUCLEUS_DATA_DIR")
        .unwrap_or_else(|_| {
            dirs::data_dir()
                .unwrap_or_else(|| PathBuf::from("/tmp"))
                .join("nucleus")
                .to_string_lossy()
                .to_string()
        });

    let data_path = PathBuf::from(&data_dir);
    std::fs::create_dir_all(&data_path).expect("Failed to create data directory");

    // Open sled database
    let db = sled::open(data_path.join("nucleus.db")).expect("Failed to open sled database");

    let redis_url =
        std::env::var("NUCLEUS_REDIS_URL").unwrap_or_else(|_| "redis://127.0.0.1:6379".to_string());

    let shell = NucleusShell::new(db, &redis_url);

    log::info!("Session ID: {}", shell.session_id());
    println!(
        "\x1b[1;32m[NUCLEUS]\x1b[0m Session started: {}",
        shell.session_id()
    );
    println!(
        "\x1b[1;32m[NUCLEUS]\x1b[0m AI-native shell runtime active. Type commands normally."
    );
    println!(
        "\x1b[1;32m[NUCLEUS]\x1b[0m Risk analysis, context tracking, and rollback enabled."
    );
    println!();

    // Check for single command execution mode: nucleus -c "command"
    if args.len() >= 3 && args[1] == "-c" {
        let command = &args[2];
        let node = shell.execute_command(command);
        print!("{}", node.stdout);
        if !node.stderr.is_empty() {
            eprint!("{}", node.stderr);
        }
        std::process::exit(node.exit_code);
    }

    // Interactive mode
    match shell.run() {
        Ok(()) => {
            log::info!("NUCLEUS session ended normally");
        }
        Err(e) => {
            log::error!("NUCLEUS session error: {}", e);
            eprintln!("\x1b[1;31m[NUCLEUS]\x1b[0m Error: {}", e);
            std::process::exit(1);
        }
    }
}
