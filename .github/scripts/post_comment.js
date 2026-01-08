module.exports = async ({github, context, core}) => {
  const fs = require('fs');
  
  // Validate pr_number.txt
  if (!fs.existsSync('pr_number.txt')) {
    core.setFailed("Required artifact file 'pr_number.txt' was not found in the workspace.");
    return;
  }
  const prNumberContent = fs.readFileSync('pr_number.txt', 'utf8').trim();
  const issue_number = parseInt(prNumberContent, 10);
  if (!Number.isFinite(issue_number) || issue_number <= 0) {
     core.setFailed('Invalid PR number in pr_number.txt: "' + prNumberContent + '"');
     return;
  }

  // Validate comparison.md
  if (!fs.existsSync('comparison.md')) {
    core.setFailed("Required artifact file 'comparison.md' was not found in the workspace.");
    return;
  }
  let comparison;
  try {
    comparison = fs.readFileSync('comparison.md', 'utf8');
  } catch (error) {
    core.setFailed("Failed to read 'comparison.md': " + error.message);
    return;
  }

  // Find existing comment
  const { data: comments } = await github.rest.issues.listComments({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: issue_number,
  });

  const botComment = comments.find(comment =>
    comment.user.type === 'Bot' &&
    comment.body.includes('Benchmark Comparison')
  );

  const footer = '<sub>🤖 This comment will be automatically updated with the latest benchmark results.</sub>';
  const commentBody = `${comparison}\n\n${footer}`;

  if (botComment) {
    await github.rest.issues.updateComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      comment_id: botComment.id,
      body: commentBody
    });
  } else {
    await github.rest.issues.createComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issue_number,
      body: commentBody
    });
  }
};
