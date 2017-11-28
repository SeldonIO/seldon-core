pipeline {
    agent none
    parameters {
        booleanParam(name: 'IS_IMAGE_BEING_PUBLISHED', defaultValue: false, description: '')
    }
    stages {
        stage('build-jar') {
            parallel {
                stage('build-jar_cluster-manager') {
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
                stage('build-image_cluster-manager') {
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
                stage('publish-image_cluster-manager') {
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
                    when {
                        expression { return params.IS_IMAGE_BEING_PUBLISHED }
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

