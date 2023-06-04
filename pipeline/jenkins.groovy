def available_os
def available_arch
def func_ChoiceOs(comm, os, arch) {
    available_os = ['linux', 'darwin', 'windows']
    if (os == 'all') {
        for (int i = 0; i < available_os.size(); ++i) {
            func_ChoiceArch(comm, available_os[i], arch)
        }
    } else {
        func_ChoiceArch(comm, os, arch)
    }
}
def func_ChoiceArch(comm, os, arch) {
    available_arch = ['amd64', 'arm', 'arm64']
    if (arch == 'all') {
        for (int i = 0; i < available_arch.size(); ++i) {
            func_DoCommand(comm, os, available_arch[i])
        }
    } else {
        func_DoCommand(comm, os, arch)
    }
}
def func_DoCommand(comm, os, arch) {
    if (os == "darwin" && arch == "arm") { return }
    echo "Do ${comm} for ${os} - ${arch}"
    sh "make ${comm} TARGETOS=${os} TARGETARCH=${arch}"
}
pipeline {
    agent any
    environment {
        REPO = 'https://github.com/NickP007/kbot'
        BRANCH = 'main'
        DIR_NAME = sh(script: 'basename "${REPO}"', returnStdout: true).trim()
    }
    /* script {
        def avail_os = ['linux', 'darwin', 'windows', 'all']
    }*/
    parameters {
        
        choice(name: 'TARGETOS', choices: ['linux', 'darwin', 'windows', 'all'], description: 'Pick OS')

        choice(name: 'TARGETARCH', choices: ['amd64', 'arm', 'arm64', 'all'], description: 'Pick ARCH')

    }
    stages {
        stage('Clone') {
            steps {
                echo "Clone Repository..."
                git branch: "${BRANCH}", url: "${REPO}"
            }
        }

        stage('Test') {
            steps {
                echo "Test ..."
                script {
                    func_ChoiceOs('test', params.TARGETOS, params.TARGETARCH)
                }
            }
        }
        stage('Build') {
            steps {
                echo "Build ..."
                script {
                    func_ChoiceOs('build', params.TARGETOS, params.TARGETARCH)
                }
            }
        }
        stage('Image') {
            steps {
                echo "Make image ..."
                script {
                    func_ChoiceOs('image', params.TARGETOS, params.TARGETARCH)
                }
            }
        }
        stage('Push') {
            steps {
                echo "Make image push ..."
                withCredentials([usernamePassword(credentialsId: 'dockerhub', usernameVariable: 'DOCKER_USERNAME', passwordVariable: 'DOCKER_PASSWORD')]) {
                    sh "docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD"
                }
                script {
                    func_ChoiceOs('push', params.TARGETOS, params.TARGETARCH)
                }
            }
        }
    }
}
