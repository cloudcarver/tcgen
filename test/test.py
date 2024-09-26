import subprocess
import sys

TCGEN_BIN="../bin/tcgen"
TESTMOD_DIR="./testmod"

p = subprocess.Popen(f"{TCGEN_BIN} -config tcgen.yaml", shell=True, stdout=sys.stdout, stderr=sys.stderr)

if p.wait() != 0:
    print("❌ Test failed")
    sys.exit(1)

print("✅ Test passed")
