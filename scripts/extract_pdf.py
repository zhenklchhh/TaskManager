import subprocess, sys

# Install PyMuPDF if needed
subprocess.check_call([sys.executable, "-m", "pip", "install", "--user", "PyMuPDF"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

import pymupdf

doc = pymupdf.open(r"c:\Users\Lenovo\projects\TaskManager\pres.pdf")
for i, page in enumerate(doc):
    print(f"=== SLIDE {i+1} ===")
    print(page.get_text())
    print()
