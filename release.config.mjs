/**
 * @type {import('semantic-release').GlobalConfig}
 */
export default {
  branches: ["release"],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    "@semantic-release/changelog",
    "@semantic-release/github",
    "@semantic-release/git",
  ],
};
