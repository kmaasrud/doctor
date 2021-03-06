import os
import subprocess
import sys
import requests
import tarfile
from math import prod
from kodb.utils import program_exists
from kodb import MSG


def download_tectonic():
    TECTONIC_VER = "0.3.3"
    print(f"Downloading tectonic AppImage {TECTONIC_VER} from GitHub...")

    if os.name == "nt":
        pass #TODO
    else:
        r = requests.get(f"https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%40{TECTONIC_VER}/tectonic-{TECTONIC_VER}-x86_64.AppImage")

        exe_path = "/usr/bin/tectonic"
        with open(exe_path, "wb") as f:
            f.write(r.content)

        mode = os.stat(exe_path).st_mode
        mode |= (mode & 0o444) >> 2
        os.chmod(exe_path, mode)

    print("Tectonic downloaded!")


def download_pandoc():
    PANDOC_VER = "2.11.3.1"
    print(f"Downloading pandoc {PANDOC_VER} from GitHub...")

    if os.name == "nt":
        pass #TODO
    else:
        r = requests.get(f"https://github.com/jgm/pandoc/releases/download/{PANDOC_VER}/pandoc-{PANDOC_VER}-linux-amd64.tar.gz")
        with open("pandoc_temp.tar.gz", "wb") as f:
            f.write(r.content)
        with tarfile.open("pandoc_temp.tar.gz", "r:gz") as f:
            f.extractall()
        os.rename(f"pandoc-{PANDOC_VER}/bin/pandoc", "/usr/bin/pandoc")
        os.system(f"rm -rf pandoc-{PANDOC_VER}")
        os.remove("pandoc_temp.tar.gz")

    print("Pandoc downloaded!")


def install_pip(package):
    subprocess.check_call([sys.executable, "-m", "pip", "install", package])


def download_dependencies(_):
    if not program_exists("pandoc"):
        MSG.warning("Pandoc does not have an automatic installation implemented  yet. Follow the installation instructions on the Pandoc website.")

    if not program_exists("tectonic"):
        MSG.warning("Tectonic does not have an automatic installation implemented  yet. Follow the installation instructions on the Tectonic website.")

    for package in ["pandoc-fignos", "pandoc-eqnos", "pandoc-tablenos", "pandoc-secnos"]:
        if not program_exists(package):
            install_pip(package)


def check_program_availability():
    for prog in ["pandoc", "tectonic", "pandoc-xnos", "pandoc-fignos", "pandoc-eqnos", "pandoc-tablenos", "pandoc-secnos"]:
        if not program_exists(prog):
            MSG.error(f"""{prog} does not exist on this system or is not in PATH. Run 'kodb --download-dependencies'
to install required dependencies (may have varying success).""")
            sys.exit()
