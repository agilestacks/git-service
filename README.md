# Git Service

This is a Git server providing access over HTTP and SSH, using `git` CLI as a repository backend on locally attached filesystem.

Git Service also has an HTTP [API] to provide access to it's functions, such as:

- creating an empty Git repository;
- creating a Git repository from template: cloning upstream (GitHub) repository or unpacking from an S3 binary blob;
- uploading files to the repository as commits;
- performing Git subtree splits to embed sources as subdirectories;
- retrieving repository log.

This software is not a standalone piece but is intended for use as part of a microservice architecture. You'd need to change authentication and authorization callouts to suit you needs. Otherwise, it's a pretty good overview how to implement a Git server in Go.

## Git Service at Agile Stacks

Git Service is to distribute Stack Templates and Stack source code. After user creates Stack Template and then shares it by setting Team(s) permissions on it - a Git repository is created. The content of repository is populated by unpacking source archive from S3. Then, new files - Hub manifests and parameters, Terraform source code, etc. - the Template specifics are added via simple file upload API.

Git Service functions called by Automation Hub:

1. Create repo and unpack archive
2. Upload file.

Name of the repository is `user's organization name/template name-template id`. Template `id` is database `templates` table primary key value.

Git Service is accessed by end-user over SSH with public key authentication. Git Service request following information from Automation Hub:

1. Get the list of users that have SSH public key equal that of received  during SSH authentication phase (the keys are offered by client). The API resource is `/user/keys?fingerprint=<public key sha256 fingerprint>`.
2. Retrieve template owner and teams permissions set on the template by extracting template `id` from accessed Git repository URL. The resource is `/templates/:id`.

Git Service requests User to Team membership information from Authentication Service (in turn backed by Okta) on `/teams/:id`.


[API]: https://agilestacks.github.io/git-service/API.html
