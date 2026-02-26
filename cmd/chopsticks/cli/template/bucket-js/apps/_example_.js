/** @type {import('./_chopsticks_.js')} */
/** @type {import('./_tools_.js')} */

/**
 * Example App
 *
 * A sample app
 * @module apps/_example_
 */

class ExampleApp extends App {
  constructor() {
    super({
      name: "example",
      description: "Example Application",
      homepage: "https://example.com",
      license: "MIT",
      bucket: "{{.Name}}",
    });
  }

  /**
   * 检查最新版本
   * @returns {Promise<string>}
   */
  async checkVersion() {
    return "1.0.0";
  }

  /**
   * 获取下载信息
   * @param {string} version
   * @param {string} arch
   * @returns {Promise<DownloadInfo>}
   */
  async getDownloadInfo(version, arch) {
    return {
      url: "https://example.com/download/" + version + "/app-" + arch + ".zip",
      type: "zip",
    };
  }

  /**
   * 安装后钩子
   * @param {InstallContext} ctx
   * @returns {Promise<void>}
   */
  async onPostInstall(ctx) {
    log.info("Installation complete");
  }
}

