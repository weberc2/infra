import "github.com/weberc2/infra"

infra.self

projects: [
    {
        name: "actionize-check"
        jobs: {
            "pull-request": [
                infra.#Job & {
                    name: "actionize-check"
                    #steps: [
                        infra.#CheckoutStep,
                        infra.#GoSetupStep,
                        {
                            name: "actionize-check"
                            run: """
                                set -eo pipefail
                                cd {{ .Path }}/apps/actionize
                                go run .
                                if [[ -n "$(git diff)" ]]; then
                                    echo "Found differences:"
                                    echo "$(git diff)"
                                    exit 1
                                fi
                                """
                        },
                    ]
                }
            ]
        }
    }
]