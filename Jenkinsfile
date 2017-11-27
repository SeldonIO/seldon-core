pipeline {
    agent none
    stages {
        stage('build-jar') {
            parallel {
                stage('cluster-manager-build-jar') {
                    agent {
                        docker {
                            image 'seldonio/core-builder'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd cluster-manager && make -f Makefile.ci clean build_jar'
                    }
                }
            }
        }
    }
}

