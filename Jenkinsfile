pipeline {
    agent {
        dockerfile {
            dir 'ci'
            args '-v ${WORKSPACE}:/usr/local/src/go/src/github.com/stardog-union/stardog-graviton -w /usr/local/src/go/src/github.com/stardog-union/stardog-graviton'
        }
    }
    environment {
        artifactoryUsername = credentials('artifactoryUsername')
        artifactoryPassword = credentials('artifactoryPassword')
        STARDOG_LICENSE = credentials('stardog_license_base64')
        AWS_ACCESS_KEY_ID = credentials('BUZZ_AWS_ACCESS_KEY_ID')
        AWS_SECRET_ACCESS_KEY = credentials('BUZZ_AWS_SECRET_ACCESS_KEY')
        GITHUB_CREDS = credentials('buzzgithub')
    }
    parameters {
        booleanParam(name: 'ACCEPTANCE_TESTS', defaultValue: true, description: 'Run the ong acceptance tests')
        string(name: 'TAG_VERSION', defaultValue: '', description: 'Run a Gradle test group')
        string(name: 'STARDOG_VERSION', defaultValue: '5.0', description: 'The version of Stardog to use with these tests')
        string(name: 'AMI', defaultValue: '', description: 'The ami to use as the base.  Advanced for debugging')
        string(name: 'MERGE_BRANCH', defaultValue: '', description: 'The branch to merge into')
        string(name: 'REMOTE_REPO', defaultValue: '', description: 'The repository to push into')
        string(name: 'S3_BUCKET', defaultValue: 'graviton-releases', description: 'The S3 bucket where artifacts will be published')
    }
    stages {
        stage('Make graviton') {
            steps {
                sh "make"
            }
        }
        stage('Unit Tests') {
            steps {
                sh "make test"
            }
        }
        stage('Build Release') {
            steps {
                sh "./ci/build-graviton.sh"
                sh "./ci/zip.sh"
            }
        }
        stage('Acceptance Tests') {
            when { expression { params.ACCEPTANCE_TESTS } }
            steps {
                sh "./ci/make-env.sh"
                script {
                    try {
                        timeout (time: 2, unit: 'HOURS') {
                        sh "./ci/start-cluster.sh"
                        }
                    }
                    catch (error) {
                        echo "Error running acceptance tests: " + error.toString()
                        sh "./ci/stop-cluster.sh"
                        throw error
                    }
                }
                script {
                    try {
                        timeout (time: 2, unit: 'HOURS') {
                        sh "./ci/create-db.sh"
                        }
                    }
                    catch (error) {
                        echo "Error running acceptance tests: " + error.toString()
                        sh "./ci/stop-cluster.sh"
                        throw error
                    }
                }
                sh "./ci/stop-cluster.sh"
            }
        }
        stage('Merge and push') {   
            when { expression { params.S3_BUCKET != '' && params.REMOTE_REPO != ''} }             
            steps {
                sh "./ci/merge.sh"
            }
        }
        stage('Publish') {                
            steps {
                sh "./ci/publish.sh"
            }
        }
    }
}