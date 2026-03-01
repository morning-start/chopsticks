/**
 * @description Distributed version control system
 * @version 2.40.0
 * @homepage https://git-scm.com
 * @license GPL-2.0-only
 * @category development
 */

const app = {
  name: "git",
  version: "2.40.0",
  architecture: {
    "64bit": {
      url: "https://example.com/git-2.40.0-64-bit.exe",
      hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  },
  bin: ["cmd\\git.exe", "cmd\\gitk.exe", "cmd\\git-gui.exe"]
};

module.exports = app;
