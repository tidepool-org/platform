@Library('mdblp-library') _
def builderImage
pipeline {
    agent any
    stages {
        stage('Initialization') {
            steps {
                script {
                    utils.initPipeline()
                    if(env.GIT_COMMIT == null) {
                        // git commit id must be a 40 characters length string (lower case or digits)
                        env.GIT_COMMIT = "f".multiply(40)
                    }
                    builderImage = docker.build('go-build-image','-f ./Dockerfile.build .')
                    env.RUN_ID = UUID.randomUUID().toString()
                    docker.image('docker.ci.diabeloop.eu/ci-toolbox').inside() {
                        env.version = sh (
                            script: 'release-helper get-version',
                            returnStdout: true
                        ).trim().toUpperCase()
                    }
                }
            }
        }
        stage('Build ') {
            steps {
                script {
                    builderImage.inside("") {
                        sh "make ci-generate ci-build"
                    }
                }
            }
        }
        stage('Test ') {
            steps {
                echo 'start mongo to serve as a testing db'
                sh """
                    docker network create platform_build${RUN_ID}

                    docker container run -d --ulimit nofile=1048576 --name mongo4platform${RUN_ID} --network=platform_build${RUN_ID} mongo:4.2

                """
                script {
                    builderImage.inside("--network=platform_build${RUN_ID}") {

                        sh "JENKINS_TEST=on make ci-test"
                    }
                }
            }
            post {
                always {
                    sh """
                        docker logs mongo4platform${RUN_ID} > mongo4platform.log

                        gzip -9f mongo4platform.log
                    """
                    archiveArtifacts artifacts: 'mongo4platform.log.gz'
                    sh 'docker stop mongo4platform${RUN_ID} && docker rm mongo4platform${RUN_ID}  && docker network rm platform_build${RUN_ID}'

                    junit '**/junit-report/report.xml'
                }
            }
        }
        stage('Package') {
            steps {
                pack()
            }
        }
        
        stage('Documentation') {
            steps {
                script {
                    builderImage.inside("") {
                        sh """
                            SERVICE=data make ci-soups
                            ./buildDoc.sh
                            mkdir -p ./ci-doc
                            mv ./soup/platform/platform-0.0.0-soup.md ./ci-doc/platform-${version}-soup.md
                            mv ./docs/api/v1/data/swagger.json ./ci-doc/platform-${version}-swagger.json

                            cp ./ci-doc/platform-${version}-swagger.json ./ci-doc/platform-latest-swagger.json
                        """
                        dir("ci-doc") {
                            stash name: "doc", includes: "*", allowEmtpy: true
                        }
                    }
                }
                
            }
        }
        stage('Publish') {
            when { branch "dblp" }
            steps {
                publish()
            }
        }
    }
}
