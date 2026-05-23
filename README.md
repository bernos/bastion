# Bastion

- Gemini conversation about doing ssh natively in go: https://share.google/aimode/8QSZyPHELBXy8O6CU

## Workflow

### Propose on the main branch

Always run `/opsx:propose`  (or `/opsx:explore` and continue) on the *main branch*, never inside a worktree branch.

The *main* branch is the source-of-truth. OpenSpec's proposal step uses the current work branch as the baseline for proposing changes. If we don't propose on the *main* branch, we're not baselining off of the source-of-truth and we may lose visibility of in-flight changes being made in parallel.

### Apply with sub agents and worktrees

Once proposals are ready on main, hand each one off to a sub agent with its own worktree. Each sub agent shouldwork on an isolated branch. Because worktrees share the same git object store, you get true parallelism without needing to duplicate the repo.

Always *verify* inside the same subagent session, before merging the branch.

### Merge first, then archive

Once a worktree has been verified, raise a PR for it, review the PR and then merge back to the main branch. Do this *before* archiving the change with `/opsx:archive`. If you archive on the worktree branch, the spec merge runs against an incomplete view of the specs on `main` - other features that have been merged in the meantime may not be visible to the branch yet.

### Commit Discipline

More commits means more recovery points. Commit at every logical boundary:

- After the change proposal is created, or updated.
- After each logical code change inside the workgree branch.
- After running Archive.

Don't be afraid to create many, many commits. When things go of track, a dense commit history gives you a clean point to go back to and continue from, rather than having to start all over.



