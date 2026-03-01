/**
 * @description A test application for integration testing
 * @version 1.0.0
 * @homepage https://github.com/test/test-app
 * @license MIT
 * @category test
 */

const app = {
  name: "test-app",
  version: "1.0.0",
  architecture: {
    "64bit": {
      url: "https://example.com/test-app-1.0.0.zip",
      hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  },
  bin: ["test-app.exe"]
};

module.exports = app;
