package infra

#Step: {
    name?: string
    uses?: string
    env?: {[string]: string}
    run?: string
}

#CheckoutStep: #Step & { uses: "actions/checkout@v2" }

#Job: {
    #steps: [...#Step]
    {
        name: string
        needs: [...string]
        "runs-on": string | *"ubuntu-latest"
        steps: [#CheckoutStep] + #steps
    }
}

#Project: {
    name: string
    jobs: ["merge" | "pull-request"]: [...#Job]
}

#GoSetupStep: #Step & {
    name: "Go setup"
    uses: "actions/setup-go@v2"
}

#GoJob: #Job & {
    #name: string
    #command: string
    #args: string
    _#steps: [...#Step]
    {
        name: #name
        #steps: [
            #GoSetupStep,
            {
                name: "Go \(#name)"
                run: "(cd {{ .Path }} && \(#command) \(#args))"
            },
        ] + _#steps
    }
}

#GoToolJob: {
    _#command: string
    _#args: string
    _#steps_: [...#Step]
    #GoJob & {
        #name: _#command
        #command: "go"
        #args: "\(_#command) \(_#args)"
        _#steps: _#steps_
    }
}

#GoTestJob: #GoToolJob & {
    _#command: "test"
    _#args: "-v ./..."
}

#GoLintJob: #Job & {
    name: "lint"
    #steps: [
        #GoSetupStep,
        {
            name: "Fetch golint"
            run: """
                (
                    cd {{ .Path }} &&
                    mkdir bin &&
                    export GOBIN=$PWD/bin &&
                    echo "GOBIN=$GOBIN" >> $GITHUB_ENV &&
                    echo "PATH=$GOBIN:$PATH" >> $GITHUB_ENV &&
                    go get golang.org/x/lint/golint
                )
                """
        },
        {
            name: "Go lint"
            // -set_exit_status is golint's dumb way of spelling `--check`
            run: "(cd {{ .Path }} && golint -set_exit_status ./...)"
        }
    ]
}

#GoBuild: #Job & {
    name: "build"
    #steps: [
        #GoSetupStep,
        {
            name: "Go build"
            run: """
            (
                cd {{ .Path }} &&
                export OUTPUT={{ .Name }} &&
                echo "OUTPUT=$OUTPUT" >> $GITHUB_ENV &&
                go build -o $OUTPUT
            )
            """
        },
        {
            name: "Zip"
            run: """
            (
                export OUTPUT_ZIP=$OUTPUT-$(git rev-parse HEAD).zip
                echo "OUTPUT_ZIP=$OUTPUT_ZIP" >> $GITHUB_ENV
                cd {{ .Path }} && zip $OUTPUT_ZIP $OUTPUT
            )
            """
        },
        {
            name: "S3 Upload"
            env: #AWSEnv
            run: """
            (
                cd {{ .Path}} &&
                aws s3 cp $OUTPUT_ZIP s3://weberc2-prd-lambda-support-code-artifacts/$OUTPUT_ZIP
            )
            """
        }
    ]
}

#GoProject: #Project & {
    #pullRequestJobs: [...#Job]
    #mergeJobs: [...#Job]
    {
        jobs: {
            "pull-request": [ #GoTestJob, #GoLintJob ] + #pullRequestJobs
            merge: [ #GoTestJob, #GoLintJob ] + #mergeJobs
        }
    }
}

#TerraformSetupStep: #Step & {
    name: "Terraform setup"
    uses: "hashicorp/setup-terraform@v1"
}

#TerraformFmtJob: #Job & {
    name: "fmt"
    #steps: [
        #TerraformSetupStep,
        {
            name: "Terraform fmt"
            run: "terraform -chdir={{ .Path }} fmt -recursive -check"
        }
    ]
}

#AWSEnv: {
    AWS_ACCESS_KEY_ID: "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}"
    AWS_SECRET_ACCESS_KEY: "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}"
    AWS_REGION: "us-east-2"
}

#TerraformStep: #Step & {
    #name: "init" | "plan" | "apply"
    {
        name: "Terraform \(#name)"
        env: #AWSEnv
        run: "terraform -chdir={{ .Path }} \(#name)"
    }
}

#TerraformJob: #Job & {
    #command: "apply" | "plan"
    {
        name: #command
        #steps: [
            #TerraformSetupStep,
            #TerraformStep & { #name: "init" },
            #TerraformStep & { #name: #command },
        ]
    }
}

#TerraformProject: #Project & {
    name: string
    jobs: {
        merge: [
            #TerraformFmtJob,
            #TerraformJob & { #command: "apply" },
        ]
        "pull-request": [
            #TerraformFmtJob,
            #TerraformJob & { #command: "plan"} ,
        ]
    }
}

constraints: {
    projects: [...#Project]
}