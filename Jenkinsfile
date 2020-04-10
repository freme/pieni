#!groovy

pipeline {
    agent { label 'dockerv2' }
    options {
        ansiColor('gnome-terminal')
        buildDiscarder logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10')
        disableConcurrentBuilds()
        skipDefaultCheckout true
        preserveStashes buildCount: 10
        retry(3)
        timeout(10)
        timestamps()
    }
    environment {
        REG = 'docker-dev-local.getgotools.net'
        APP = 'eks/pieni'
        VER = 'latest'
        IMG = "${REG}/${APP}:${VER}"
        OWN = 'thomas.william@logmein.com'
        TNM = 'RTC Productivity Dresden'
    }
    libraries {
        lib('pipeline-library@master')
    }

    stages {
        stage('DockerImageCreation') {
            steps {
                cleanWs notFailBuild: true
                checkout scm
                script {
                    def img = docker.build(env.IMG)
                    docker.withRegistry("https://${env.REG}", 'DockerCredentialsForArtifactory') {
                        img.push()
                    }
                }
            }
        }
        stage('Benchmark') {
            steps {
                sh "docker tag ${env.IMG} pieni"
                sh './benchmark.sh'
            }
        }
        stage('DockerImagePublish') {
            steps {
                script {
                    tPromoteDocker(imagePath: env.APP, imageTag: env.VER, owner: env.OWN, teamName: env.TNM)
                    echo 'To pull the new image, do:'
                    echo "docker pull docker-release.getgotools.net/${env.APP}:${env.VER}"
                }
            }
        }
    }
}
