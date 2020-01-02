.ONESHELL:

all: base loanclassifier outliersdetector combiner

base:
	docker build . -t seldon-core-outliers-base:0.1

loanclassifier:
	s2i build pipeline/loanclassifier seldon-core-outliers-base:0.1 loanclassifier:0.1

outliersdetector:
	s2i build pipeline/outliersdetector seldon-core-outliers-base:0.1 outliersdetector:0.1

combiner:
	s2i build pipeline/combiner seldon-core-outliers-base:0.1 combiner:0.1
