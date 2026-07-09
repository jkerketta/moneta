from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

PLIST_DIR = Path.home() / "Library/LaunchAgents"
PLIST_NAME = "com.moneta.scan"
PLIST_PATH = PLIST_DIR / f"{PLIST_NAME}.plist"

PLIST_TEMPLATE = """<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{name}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{python}</string>
        <string>-m</string>
        <string>moneta</string>
        <string>scan</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>8</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    <key>WorkingDirectory</key>
    <string>{cwd}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>FINNHUB_API_KEY</key>
        <string>{finnhub_key}</string>
    </dict>
    <key>StandardOutPath</key>
    <string>{log_dir}/moneta-stdout.log</string>
    <key>StandardErrorPath</key>
    <string>{log_dir}/moneta-stderr.log</string>
</dict>
</plist>
"""


def install_plist() -> None:
    from moneta.config import load_env
    load_env()

    PLIST_DIR.mkdir(parents=True, exist_ok=True)
    log_dir = Path.home() / ".moneta" / "logs"
    log_dir.mkdir(parents=True, exist_ok=True)

    content = PLIST_TEMPLATE.format(
        name=PLIST_NAME,
        python=sys.executable,
        cwd=str(Path.cwd()),
        finnhub_key=os.environ.get("FINNHUB_API_KEY", ""),
        log_dir=str(log_dir),
    )

    PLIST_PATH.write_text(content, encoding="utf-8")
    subprocess.run(["launchctl", "load", str(PLIST_PATH)], check=True)


def uninstall_plist() -> None:
    if PLIST_PATH.exists():
        subprocess.run(["launchctl", "unload", str(PLIST_PATH)], check=True)
        PLIST_PATH.unlink()


def status() -> str:
    if not PLIST_PATH.exists():
        return "No scheduled scan configured."
    result = subprocess.run(
        ["launchctl", "list", PLIST_NAME],
        capture_output=True, text=True,
    )
    if result.returncode == 0:
        return "Scheduled scan active (runs daily at 8:00 AM via launchd)."
    return "Plist exists but not loaded."
