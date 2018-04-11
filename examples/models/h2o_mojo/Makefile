

build_jar: 
	mvn clean package 

clean:
	mvn clean
	rm -rf experiment
	rm -f src/main/resources/model.zip
train:
	mkdir -p experiment
	python train.py
	mv experiment/*.zip src/main/resources/model.zip
