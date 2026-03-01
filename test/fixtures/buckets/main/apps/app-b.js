/**
 * @description App B - depends on App C
 * @version 1.0.0
 * @category test
 */

const app = {
  name: "app-b",
  version: "1.0.0",
  depends: ["app-c"],
  architecture: {
    "64bit": {
      url: "https://example.com/app-b-1.0.0.zip",
      hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  },
  bin: ["app-b.exe"]
};

module.exports = app;
