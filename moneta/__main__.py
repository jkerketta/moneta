import sys

from moneta.app import MonetaApp
from moneta.cli import app as cli_app


def main():
    if len(sys.argv) > 1:
        cli_app()
    else:
        tui = MonetaApp()
        tui.run()


if __name__ == "__main__":
    main()
