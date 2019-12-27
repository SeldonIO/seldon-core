.ONESHELL:

all: loanclassifier outliersdetector

loanclassifier:
	s2i build pipeline/loanclassifier seldonio/seldon-core-s2i-python37:0.14 loanclassifier:0.1

outliersdetector:
	s2i build pipeline/outliersdetector seldonio/seldon-core-s2i-python37:0.14 outliersdetector:0.1
