# Security Policy

## Reporting a vulnerability

Do not open a public issue for suspected vulnerabilities or exposed credentials.
Use GitHub's private vulnerability reporting for this repository instead.

Include a clear description, reproduction steps, affected versions, and the
potential impact. Please do not include real credentials, webhook URLs, phone
numbers, spreadsheet IDs, or personal data in the report.

## Credential handling

Service-account JSON files, `.env` files, kubeconfigs, webhook credentials, and
recipient identifiers must remain outside the repository and container image.
If a secret is committed, revoke or rotate it immediately; deleting the file in
a later commit is not sufficient because Git retains history.
