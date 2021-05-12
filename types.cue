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

#GoLintJob: #GoJob & {
    #name: "lint"
    #command: "go"
    #args: "get golang.org/x/lint/golint"
    _#steps: [
        {
            name: "Go lint"
            run: "(cd {{ .Path }} && golint ./...)"
        }
    ]
}

#GoProject: #Project & {
    jobs: {
        "pull-request": [ #GoTestJob, #GoLintJob ]
        merge: [ #GoTestJob, #GoLintJob ]
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

#TerraformStep: #Step & {
    #name: "init" | "plan" | "apply"
    {
        name: "Terraform \(#name)"
        env: {
            AWS_ACCESS_KEY_ID: "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}"
            AWS_SECRET_ACCESS_KEY: "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}"
        }
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