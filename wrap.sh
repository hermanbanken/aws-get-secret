# Get the secret value by calling the Go executable
values=$(${fullPath}/go-retrieve-secret -r "${region}" -s "${secretArn}" -a "${roleName}" -t ${timeout})
last_cmd=$?

# Verify that the last command was successful
if [[ ${last_cmd} -ne 0 ]]; then
    echo "Failed to setup environment for Secret ${secretArn}"
    exit 1
fi