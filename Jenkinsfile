pipeline {
    agent none
    parameters {
        booleanParam(name: 'IS_IMAGE_BEING_PUBLISHED', defaultValue: false, description: '')
    }
    environment {
        SELDON_CORE_DOCKER_HUB_USER = credentials('SELDON_CORE_DOCKER_HUB_USER')
        SELDON_CORE_DOCKER_HUB_PASSWORD = credentials('SELDON_CORE_DOCKER_HUB_PASSWORD')
    }
    stages {
        stage('build-api-frontend') {
            agent {
                docker {
                    image 'seldonio/core-builder:0.1'
                    args '-v /root/.m2:/root/.m2'
                }
            }
            steps {
                echo "Build Jar"
                sh 'cd api-frontend && make -f Makefile.ci clean build_jar'
                echo "Build Image"
                sh 'cd api-frontend && make -f Makefile.ci write_version'
                sh 'cd api-frontend && make -f Makefile.ci build_image'
                script {
                    if (params.IS_IMAGE_BEING_PUBLISHED == true) {
                        echo "Publish Image"
                        sh 'cd api-frontend && make -f Makefile.ci repo_login'
                        sh 'cd api-frontend && make -f Makefile.ci push_image'
                    } else {
                        echo "Publish Image [SKIPPED]"
                    }
                }
            }
        }
        stage('build-engine') {
            agent {
                docker {
                    image 'seldonio/core-builder:0.1'
                    args '-v /root/.m2:/root/.m2'
                }
            }
            steps {
                echo "Build Jar"
                sh 'cd engine && make -f Makefile.ci clean build_jar'
                echo "Build Image"
                sh 'cd engine && make -f Makefile.ci write_version'
                sh 'cd engine && make -f Makefile.ci build_image'
                script {
                    if (params.IS_IMAGE_BEING_PUBLISHED == true) {
                        echo "Publish Image"
                        sh 'cd engine && make -f Makefile.ci repo_login'
                        sh 'cd engine && make -f Makefile.ci push_image'
                    } else {
                        echo "Publish Image [SKIPPED]"
                    }
                }
            }
        }
        stage('build-cluster-manager') {
            agent {
                docker {
                    image 'seldonio/core-builder:0.1'
                    args '-v /root/.m2:/root/.m2'
                }
            }
            steps {
                echo "Build Jar"
                sh 'cd cluster-manager && make -f Makefile.ci clean build_jar'
                echo "Build Image"
                sh 'cd cluster-manager && make -f Makefile.ci write_version'
                sh 'cd cluster-manager && make -f Makefile.ci build_image'
                script {
                    if (params.IS_IMAGE_BEING_PUBLISHED == true) {
                        echo "Publish Image"
                        sh 'cd cluster-manager && make -f Makefile.ci repo_login'
                        sh 'cd cluster-manager && make -f Makefile.ci push_image'
                    } else {
                        echo "Publish Image [SKIPPED]"
                    }
                }
            }
        }
    }
}

