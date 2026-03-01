/**
 * @description App C - no dependencies
 * @version 1.0.0
 * @category test
 */

const app = {
  name: "app-c",
  version: "1.0.0",
  depends: [],
  architecture: {
    "64bit": {
      url: "https://example.com/app-c-1.0.0.zip",
      hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  },
  bin: ["app-c.exe"]
};

module.exports = app;
