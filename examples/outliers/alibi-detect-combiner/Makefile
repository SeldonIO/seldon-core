.ONESHELL:

all: loanclassifier outliersdetector

loanclassifier:
	s2i build pipeline/loanclassifier seldonio/seldon-core-s2i-python37:0.14 loanclassifier:0.1

outliersdetector:
	s2i build pipeline/outliersdetector seldonio/seldon-core-s2i-python37:0.14 outliersdetector:0.1

combiner:
	s2i build pipeline/combiner seldonio/seldon-core-s2i-python37:0.14 combiner:0.1
