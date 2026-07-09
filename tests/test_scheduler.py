from __future__ import annotations

from pathlib import Path
from unittest.mock import patch
import pytest
from moneta.scheduler import install_plist, uninstall_plist, status


def test_install_plist_creates_file(tmp_path: Path):
    plist_file = tmp_path / "com.moneta.scan.plist"

    with patch("moneta.scheduler.PLIST_PATH", plist_file), \
         patch("moneta.scheduler.subprocess.run") as mock_run, \
         patch("moneta.config.load_env"), \
         patch("moneta.scheduler.Path.home", return_value=tmp_path), \
         patch("moneta.scheduler.Path.cwd", return_value=tmp_path):
        install_plist()
        assert plist_file.exists()
        content = plist_file.read_text()
        assert "com.moneta.scan" in content
        assert "8</integer>" in content
        assert "moneta" in content
        mock_run.assert_called_once()


def test_uninstall_plist_removes_existing_file(tmp_path: Path):
    plist_dir = tmp_path / "LaunchAgents"
    plist_dir.mkdir(parents=True)
    plist_file = plist_dir / "com.moneta.scan.plist"
    plist_file.write_text("test")

    with patch("moneta.scheduler.PLIST_PATH", plist_file), \
         patch("moneta.scheduler.subprocess.run") as mock_run:
        uninstall_plist()
        assert not plist_file.exists()
        assert mock_run.call_count >= 1


def test_uninstall_plist_no_file(tmp_path: Path):
    plist_file = tmp_path / "nonexistent.plist"

    with patch("moneta.scheduler.PLIST_PATH", plist_file), \
         patch("moneta.scheduler.subprocess.run") as mock_run:
        uninstall_plist()
        mock_run.assert_not_called()


def test_status_not_configured(tmp_path: Path):
    plist_file = tmp_path / "com.moneta.scan.plist"

    with patch("moneta.scheduler.PLIST_PATH", plist_file):
        result = status()
        assert "No scheduled scan configured" in result


def test_status_active(tmp_path: Path):
    plist_file = tmp_path / "com.moneta.scan.plist"
    plist_file.write_text("test")

    with patch("moneta.scheduler.PLIST_PATH", plist_file), \
         patch("moneta.scheduler.subprocess.run") as mock_run:
        mock_run.return_value.returncode = 0
        mock_run.return_value.stdout = "PID\tStatus\tLabel\n12345\t0\tcom.moneta.scan\n"
        result = status()
        assert "active" in result.lower() or "8:00" in result


def test_status_plist_exists_not_loaded(tmp_path: Path):
    plist_file = tmp_path / "com.moneta.scan.plist"
    plist_file.write_text("test")

    with patch("moneta.scheduler.PLIST_PATH", plist_file), \
         patch("moneta.scheduler.subprocess.run") as mock_run:
        mock_run.return_value.returncode = 1
        result = status()
        assert "Plist exists but not loaded" in result
