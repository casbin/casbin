module.exports = async ({github, context, core}) => {
  try {
    const artifacts = await github.rest.actions.listWorkflowRunArtifacts({
      owner: context.repo.owner,
      repo: context.repo.repo,
      run_id: context.payload.workflow_run.id,
    });

    const matchArtifact = artifacts.data.artifacts.find((artifact) => {
      return artifact.name == "benchmark-results";
    });

    if (!matchArtifact) {
      core.setFailed("No artifact named 'benchmark-results' found.");
      return;
    }

    const download = await github.rest.actions.downloadArtifact({
      owner: context.repo.owner,
      repo: context.repo.repo,
      artifact_id: matchArtifact.id,
      archive_format: 'zip',
    });

    const fs = require('fs');
    const path = require('path');
    const workspace = process.env.GITHUB_WORKSPACE;
    fs.writeFileSync(path.join(workspace, 'benchmark-results.zip'), Buffer.from(download.data));
  } catch (error) {
    core.setFailed(`Failed to download artifact: ${error.message}`);
  }
};
