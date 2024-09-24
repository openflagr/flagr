// Wiki: https://acikota.atlassian.net/wiki/spaces/Platform/pages/311787592/Centralising+stages+of+the+pipeline+for+golang+backend+services
@Library('Allen_Shared_Libraries') _
commonPipelineForGolang(
    jenkinsAgentImage: '537984406465.dkr.ecr.ap-south-1.amazonaws.com/allen-jenkins-agent:v4',
    environment: [
        GOPRIVATE : "https://github.com/Allen-Career-Institute/*",
        DEPLOY_GITOPS_REPO : "central-gitops-repo",
        DEPLOY_TARGET_FILE : "app-charts/flagr/values-dev.yaml",
        DEPLOY_TARGET_FILE_LIVE : "app-charts-live/flagr/values-dev.yaml",
        DEPLOY_TARGET_BRANCH : "main",
        REGISTRY : "537984406465.dkr.ecr.ap-south-1.amazonaws.com",
        REPOSITORY : "flagr",
        SLACK_CHANNEL : "pipeline-comms"
    ]
)
