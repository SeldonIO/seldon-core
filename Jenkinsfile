pipeline {
    parameters {
        booleanParam(name: 'IS_DISABLED_APIFE_BUILD', defaultValue: false, description: '')
        booleanParam(name: 'IS_DISABLED_ENGINE_BUILD', defaultValue: false, description: '')
        booleanParam(name: 'IS_IMAGE_BEING_PUBLISHED', defaultValue: false, description: '')
    }
    agent {
        docker {
            image 'seldonio/core-builder:0.8'
            args '-v /root/.m2:/root/.m2'
        }
    }
    environment {
        SELDON_CORE_DOCKER_HUB_USER = credentials('SELDON_CORE_DOCKER_HUB_USER')
        SELDON_CORE_DOCKER_HUB_PASSWORD = credentials('SELDON_CORE_DOCKER_HUB_PASSWORD')
    }
    stages {
        stage('build-engine') {
            when {
                expression { return params.IS_DISABLED_ENGINE_BUILD == false }
            }
            steps {
                echo "Build Image"
                sh 'cd engine && make -f Makefile.ci build'
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
    }
}

