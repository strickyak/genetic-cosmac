from go import runtime
from go import github.com/strickyak/genetic-cosmac as gene

def main(argv):
  n = runtime.NumCPU()
  say n
  runtime.GOMAXPROCS(n)
  say gene.RunN(n)
