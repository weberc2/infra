projects: [{
    name: "bootstrap-tf"
    jobs: "pull-request": [
        {
            name: "lint"
            "runs-on": "ubuntu-latest"
            steps: [
                { uses: "actions/checkout@v2" },
                {
                    name: "Terraform setup"
                    uses: "hashicorp/setup-terraform@v1"
                },
                {
                    name: "lint-terraform"
                    run: "terraform -chdir={{ .Path }} fmt -recursive -check"
                },
            ]
        },
        {
            name: "plan"
            "runs-on": "ubuntu-latest"
            steps: [
                {uses: "actions/checkout@v2"},
                {
                    name: "Terraform setup"
                    uses: "hashicorp/setup-terraform@v1"
                },
                {
                    name: "Terraform init"
                    env: {
                        AWS_ACCESS_KEY_ID: "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}"
                        AWS_SECRET_ACCESS_KEY: "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}"
                    }
                    run: "terraform -chdir={{ .Path }} init"
                },
                {
                    name: "Terraform plan"
                    env: {
                        AWS_ACCESS_KEY_ID: "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}"
                        AWS_SECRET_ACCESS_KEY: "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}"
                    }
                    run: "terraform -chdir={{ .Path }} plan"
                },
            ]
        }
    ]
}]