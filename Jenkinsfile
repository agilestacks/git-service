#!/usr/bin/env groovy

import com.amazonaws.util.*
import com.amazonaws.auth.*
import com.amazonaws.services.ecr.*
import com.amazonaws.services.ecr.model.*
import com.amazonaws.regions.*
import io.fabric8.kubernetes.api.model.NamespaceBuilder
import io.fabric8.kubernetes.client.*

def secretsServiceImage
def secretsServiceEndpoint
def commit
def imageTag1
def imageTag2
def namespace = 'automation-hub'
def region    = Regions.currentRegion.name

@NonCPS
def kubeClient() {
    return new DefaultKubernetesClient(new ConfigBuilder().build())
}

@NonCPS
def authenticateECR(region=EC2MetadataUtils.getEC2InstanceRegion()){
 def token = AmazonECRClientBuilder.
               standard().
               withCredentials(InstanceProfileCredentialsProvider.getInstance()).
               withRegion(region).
               build().
               getAuthorizationToken(new GetAuthorizationTokenRequest())
 def auth = token.authorizationData[0]
 login = new String(auth.authorizationToken.decodeBase64()).tokenize(':')
 return  [
     "registry" : auth.proxyEndpoint,
     "user"     : login[0],
     "password" : login[1]
 ]
}

@NonCPS
def updateDeploymentImage(deployment, imageWithVersion) {
    def imageWithoutVersion = imageWithVersion.split(':')[0]
    deployment.spec.template.spec.containers.find({ it ->
        it.image.startsWith(imageWithoutVersion)
    }).each({ it ->
        it.image = imageWithVersion
    })
    return deployment
}

@NonCPS
def commitHash() {
  sh 'git rev-parse HEAD > commit'
  readFile('commit').trim().substring(0, 7)
}

node('master') {
    stage('Checkout') {
        // checkout scm
        git credentialsId: 'github-user', url: 'https://github.com/agilestacks/secrets-service.git'
        commit = commitHash()
    }
}

podTemplate( inheritFrom: 'agilestacks',label: 'pod',
    containers: [
      containerTemplate(name: 'go', image: 'golang:latest', ttyEnabled: true, command: 'cat')
    ],
    volumes: [emptyDirVolume(memory: false, mountPath: '/var/lib/docker')]
) {
    node('pod') {
        dir('api') {
            stage('Compile') {
                container('node') {
                    sh 'make install'       
                }
            }
        }
    }
}

