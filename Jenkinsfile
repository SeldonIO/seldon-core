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
                            image 'seldonio/core-builder:0.1'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd cluster-manager && make -f Makefile.ci clean build_jar'
                    }
                }
                stage('build-jar_engine') {
                    agent {
                        docker {
                            image 'seldonio/core-builder:0.1'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd engine && make -f Makefile.ci clean build_jar'
                    }
                }
                stage('build-jar_api-frontend') {
                    agent {
                        docker {
                            image 'seldonio/core-builder:0.1'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd api-frontend && make -f Makefile.ci clean build_jar'
                    }
                }
            }
        }
        stage('build-image') {
            parallel {
                stage('build-image_cluster-manager') {
                    agent {
                        docker {
                            image 'seldonio/core-builder:0.1'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd cluster-manager && make -f Makefile.ci write_version'
                        sh 'cd cluster-manager && make -f Makefile.ci build_image'
                    }
                }
                stage('build-image_engine') {
                    agent {
                        docker {
                            image 'seldonio/core-builder:0.1'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd engine && make -f Makefile.ci write_version'
                        sh 'cd engine && make -f Makefile.ci build_image'
                    }
                }
            }
        }
        stage('publish-image') {
            when {
                expression { return params.IS_IMAGE_BEING_PUBLISHED }
            }
            parallel {
                stage('publish-image_cluster-manager') {
                    environment {
                        SELDON_CORE_DOCKER_HUB_USER = credentials('SELDON_CORE_DOCKER_HUB_USER')
                        SELDON_CORE_DOCKER_HUB_PASSWORD = credentials('SELDON_CORE_DOCKER_HUB_PASSWORD')
                    }
                    agent {
                        docker {
                            image 'seldonio/core-builder:0.1'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd cluster-manager && make -f Makefile.ci repo_login'
                        sh 'cd cluster-manager && make -f Makefile.ci push_image'
                    }
                }
                stage('publish-image_engine') {
                    environment {
                        SELDON_CORE_DOCKER_HUB_USER = credentials('SELDON_CORE_DOCKER_HUB_USER')
                        SELDON_CORE_DOCKER_HUB_PASSWORD = credentials('SELDON_CORE_DOCKER_HUB_PASSWORD')
                    }
                    agent {
                        docker {
                            image 'seldonio/core-builder:0.1'
                            args '-v /root/.m2:/root/.m2'
                        }
                    }
                    steps {
                        sh 'cd engine && make -f Makefile.ci repo_login'
                        sh 'cd engine && make -f Makefile.ci push_image'
                    }
                }
            }
        }
    }
}

