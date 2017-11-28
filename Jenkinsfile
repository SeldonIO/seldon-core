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
        stage('build-image') {
            parallel {
                stage('cluster-manager-build-image') {
                    agent {
                        docker {
                            image 'seldonio/core-builder'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd cluster-manager && make -f Makefile.ci write_version'
                        sh 'cd cluster-manager && make -f Makefile.ci build_image'
                    }
                }
            }
        }
        stage('publish-image') {
            parallel {
                stage('cluster-manager-publish-image') {
                    environment {
                        SELDON_CORE_DOCKER_HUB_USER = credentials('SELDON_CORE_DOCKER_HUB_USER')
                        SELDON_CORE_DOCKER_HUB_PASSWORD = credentials('SELDON_CORE_DOCKER_HUB_PASSWORD')
                    }
                    agent {
                        docker {
                            image 'seldonio/core-builder'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd cluster-manager && make -f Makefile.ci repo_login'
                        sh 'cd cluster-manager && make -f Makefile.ci push_image'
                    }
                }
            }
        }
    }
}

