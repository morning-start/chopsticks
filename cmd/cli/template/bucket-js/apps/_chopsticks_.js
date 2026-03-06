/**
 * Chopsticks App 类型定义
 * @module @chopsticks/core
 * @version 1.0.0
 * @description JS 引擎对外接口类型定义，用于 Bucket 插件开发
 *
 * 变更日志：
 * - 1.0.0: 初始版本，定义核心类型和全局对象
 */

/**
 * Chopsticks API 版本号
 * @constant {string}
 */
const CHOPSTICKS_API_VERSION = "1.0.0";

/**
 * @typedef {Object} FsResult
 * @property {boolean} success
 * @property {string} [data]
 * @property {string} [error]
 */

/**
 * @typedef {Object} AppMetadata
 * @property {string} name
 * @property {string} [description]
 * @property {string} [homepage]
 * @property {string} [license]
 * @property {string} bucket
 */

/**
 * @typedef {'zip'|'7z'|'tar'|'tar.gz'|'tar.xz'|'exe'|'msi'} DownloadType
 */

/**
 * @typedef {Object} DownloadInfo
 * @property {string} url
 * @property {DownloadType} type
 * @property {string} [filename]
 * @property {Object} [checksum]
 * @property {'sha256'|'md5'} checksum.type
 * @property {string} checksum.value
 */

/**
 * @typedef {'amd64'|'x86'|'arm64'} Architecture
 */

/**
 * @typedef {Object} InstallContext
 * @property {string} name
 * @property {string} version
 * @property {Architecture} arch
 * @property {string} cookDir
 * @property {string} bucket
 * @property {string} downloadPath
 */

/**
 * @typedef {Object} Log
 * @property {function(string): void} debug
 * @property {function(string): void} info
 * @property {function(string): void} warn
 * @property {function(string): void} error
 */

/**
 * @typedef {Object} PathResult
 * @property {boolean} success
 * @property {string} [path]
 * @property {string} [dir]
 * @property {string} [name]
 * @property {string} [ext]
 * @property {boolean} [exists]
 * @property {boolean} [isDir]
 * @property {boolean} [isAbs]
 * @property {string} [error]
 */

/**
 * @typedef {Object} Pathx
 * @property {function(...string): PathResult} join
 * @property {function(string): PathResult} abs
 * @property {function(string): PathResult} dir
 * @property {function(string): PathResult} base
 * @property {function(string): PathResult} ext
 * @property {function(string): PathResult} clean
 * @property {function(string): PathResult} isAbs
 * @property {function(string): PathResult} exists
 * @property {function(string): PathResult} isDir
 */

/**
 * @typedef {Object} ExecResult
 * @property {boolean} success
 * @property {number} exitCode
 * @property {string} stdout
 * @property {string} stderr
 * @property {string} [error]
 */

/**
 * @typedef {Object} ExecOptions
 * @property {string} [cwd]
 * @property {Object.<string, string>} [env]
 * @property {number} [timeout]
 */

/**
 * @typedef {Object} Execx
 * @property {function(string, ...string|ExecOptions): Promise<ExecResult>} exec
 * @property {function(string, ExecOptions=): Promise<ExecResult>} shell
 * @property {function(string, ExecOptions=): Promise<ExecResult>} powershell
 */

/**
 * @typedef {Object} FetchResponse
 * @property {boolean} success
 * @property {number} status
 * @property {boolean} ok
 * @property {string} body
 * @property {Object.<string, string>} headers
 * @property {string} [error]
 */

/**
 * @typedef {Object} FetchOptions
 * @property {string} [method]
 * @property {Object.<string, string>} [headers]
 * @property {*} [body]
 * @property {string} [contentType]
 * @property {number} [timeout]
 */

/**
 * @typedef {Object} FetchClient
 * @property {function(string, FetchOptions=): FetchResponse} get
 * @property {function(string, *, string=, FetchOptions=): FetchResponse} post
 */

/**
 * @typedef {Object} UrlInfo
 * @property {boolean} success
 * @property {string} scheme
 * @property {string} host
 * @property {string} path
 * @property {string} query
 * @property {string} fragment
 * @property {string} [error]
 */

/**
 * @typedef {Object} Fetch
 * @property {function(string, string, Object=): {success: boolean, error?: string}} download
 * @property {function(string, FetchOptions=): FetchResponse} get
 * @property {function(string, *, string=, FetchOptions=): FetchResponse} post
 * @property {function(string, string, FetchOptions=): FetchResponse} request
 * @property {function(string, string, Object=): {success: boolean, error?: string}} downloadFile
 * @property {function(string): UrlInfo} parseURL
 * @property {function(string, Object.<string, string>): {success: boolean, url: string, error?: string}} buildURL
 * @property {function(string, FetchOptions=): {success: boolean, data: *, error?: string}} getJSON
 * @property {function(string, *, FetchOptions=): {success: boolean, data: *, error?: string}} postJSON
 * @property {function(FetchOptions=): FetchClient} newClient
 * @property {function(number): {success: boolean, error?: string}} setDefaultTimeout
 */

/**
 * @typedef {Object} FsStat
 * @property {boolean} success
 * @property {number} size
 * @property {boolean} isDirectory
 * @property {boolean} isFile
 * @property {string} mtime
 * @property {string} [error]
 */

/**
 * @typedef {Object} FsDirEntry
 * @property {boolean} success
 * @property {string[]} entries
 * @property {string} [error]
 */

/**
 * @typedef {Object} Fsutil
 * @property {function(string): FsResult} readFile
 * @property {function(string, string): FsResult} writeFile
 * @property {function(string, string): FsResult} append
 * @property {function(string): FsResult} mkdir
 * @property {function(string): FsResult} rmdir
 * @property {function(string): FsResult} remove
 * @property {function(string): FsResult} removeAll
 * @property {function(string): FsResult} mkdirAll
 * @property {function(string): FsResult} copy
 * @property {function(string): {success: boolean, exists: boolean, error?: string}} exists
 * @property {function(string): {success: boolean, isdir: boolean, error?: string}} isdir
 * @property {function(string): {success: boolean, isFile: boolean, error?: string}} isFile
 * @property {function(string): FsDirEntry} readDir
 * @property {function(string): FsStat} stat
 */

/**
 * @typedef {Object} ArchiveFileInfo
 * @property {string} name
 * @property {number} size
 * @property {boolean} isDir
 */

/**
 * @typedef {Object} ArchiveExtractResult
 * @property {boolean} success
 * @property {string[]} [extractedFiles]
 * @property {string} [error]
 */

/**
 * @typedef {Object} ArchiveListResult
 * @property {boolean} success
 * @property {ArchiveFileInfo[]} [files]
 * @property {string} [error]
 */

/**
 * @typedef {'zip'|'7z'|'tar'|'tar.gz'|'tar.xz'} ArchiveType
 */

/**
 * @typedef {Object} Archive
 * @property {function(string, string): ArchiveExtractResult} extract
 * @property {function(string, string): ArchiveExtractResult} extractZip
 * @property {function(string, string): ArchiveExtractResult} extract7z
 * @property {function(string, string): ArchiveExtractResult} extractTar
 * @property {function(string, string): ArchiveExtractResult} extractTarGz
 * @property {function(string): ArchiveListResult} list
 * @property {function(string): {success: boolean, type: ArchiveType}} detectType
 * @property {function(string): {success: boolean, isArchive: boolean}} isArchive
 */

/**
 * @typedef {Object} SymlinkResult
 * @property {boolean} success
 * @property {string} [target]
 * @property {boolean} [isSymlink]
 * @property {string} [error]
 */

/**
 * @typedef {Object} Symlink
 * @property {function(string, string): SymlinkResult} create
 * @property {function(string, string): SymlinkResult} createDir
 * @property {function(string, string): SymlinkResult} createHard
 * @property {function(string, string): SymlinkResult} createJunction
 * @property {function(string): SymlinkResult} read
 * @property {function(string): SymlinkResult} is
 * @property {function(string): SymlinkResult} remove
 */

/**
 * @typedef {Object} RegistryValueInfo
 * @property {string} name
 * @property {string|number} value
 * @property {string} type
 */

/**
 * @typedef {Object} RegistryResult
 * @property {boolean} success
 * @property {*} [value]
 * @property {string} [type]
 * @property {string} [error]
 */

/**
 * @typedef {Object} Registry
 * @property {function(string, string, string): RegistryResult} setValue
 * @property {function(string, string, number): RegistryResult} setDword
 * @property {function(string, string): RegistryResult} getValue
 * @property {function(string, string): RegistryResult} getDword
 * @property {function(string, string): RegistryResult} deleteValue
 * @property {function(string): RegistryResult} createKey
 * @property {function(string): RegistryResult} deleteKey
 * @property {function(string): {success: boolean, exists: boolean, error?: string}} keyExists
 * @property {function(string): {success: boolean, keys: string[], error?: string}} listKeys
 * @property {function(string): {success: boolean, values: RegistryValueInfo[], error?: string}} listValues
 */

/**
 * @typedef {'nsis'|'msi'|'inno'|'unknown'} InstallerType
 */

/**
 * @typedef {Object} InstallerOptions
 * @property {string} [installDir]
 * @property {boolean} [silent]
 */

/**
 * @typedef {Object} InstallerResult
 * @property {boolean} success
 * @property {number} exitCode
 * @property {string} [error]
 */

/**
 * @typedef {Object} Installer
 * @property {function(string, InstallerOptions=): InstallerResult} run
 * @property {function(string, InstallerOptions=): InstallerResult} runNSIS
 * @property {function(string, InstallerOptions=): InstallerResult} runMSI
 * @property {function(string, InstallerOptions=): InstallerResult} runInno
 * @property {function(string): {success: boolean, type: InstallerType, error?: string}} detectType
 */

/**
 * @typedef {Object} ShortcutOptions
 * @property {string} source
 * @property {string} name
 * @property {string} [description]
 * @property {string} [icon]
 * @property {string} [workingDir]
 * @property {string} [arguments]
 */

/**
 * @typedef {Object} ChopsticksResult
 * @property {boolean} success
 * @property {string} [path]
 * @property {string} [version]
 * @property {string} [value]
 * @property {string} [shimPath]
 * @property {string} [shortcutPath]
 * @property {string[]} [paths]
 * @property {string} [error]
 */

/**
 * @typedef {Object} Chopsticks
 * @property {function(string, string=): ChopsticksResult} getCookDir
 * @property {function(string): {success: boolean, version: string, error?: string}} getCurrentVersion
 * @property {function(string, string=): ChopsticksResult} addToPath
 * @property {function(string, string=): ChopsticksResult} removeFromPath
 * @property {function(string, string, string=): ChopsticksResult} setEnv
 * @property {function(string): ChopsticksResult} getEnv
 * @property {function(string, string=): ChopsticksResult} deleteEnv
 * @property {function(string, string): {success: boolean, shimPath: string, error?: string}} createShim
 * @property {function(string): ChopsticksResult} removeShim
 * @property {function(string, string[]): ChopsticksResult} persistData
 * @property {function(ShortcutOptions): {success: boolean, shortcutPath: string, error?: string}} createShortcut
 * @property {function(): ChopsticksResult} getCacheDir
 * @property {function(): ChopsticksResult} getConfigDir
 * @property {function(): ChopsticksResult} getPath
 * @property {function(): ChopsticksResult} getShimDir
 * @property {function(): ChopsticksResult} getPersistDir
 */

/**
 * @typedef {Object} JsonStringifyResult
 * @property {boolean} success
 * @property {string} [json]
 * @property {string} [error]
 */

/**
 * @typedef {Object} JsonParseResult
 * @property {boolean} success
 * @property {*} [data]
 * @property {string} [error]
 */

/**
 * @typedef {Object} Jsonx
 * @property {function(Object, string|number=): JsonStringifyResult} stringify
 * @property {function(string): JsonParseResult} parse
 */

/**
 * @typedef {Object} SemverParsed
 * @property {boolean} success
 * @property {string} [raw]
 * @property {string} [normalized]
 * @property {string} [type]
 * @property {number[]} [segments]
 * @property {string} [prerelease]
 * @property {number} [prereleaseNum]
 * @property {string} [build]
 * @property {boolean} [comparable]
 * @property {string} [error]
 */

/**
 * @typedef {Object} SemverConstraint
 * @property {boolean} success
 * @property {string} [type]
 * @property {string} [version]
 * @property {string} [operator]
 * @property {string} [error]
 */

/**
 * @typedef {Object} Semver
 * @property {function(string): SemverParsed} parse
 * @property {function(string, string): {success: boolean, result: -1|0|1, error?: string}} compare
 * @property {function(string, string): {success: boolean, result: boolean, error?: string}} gt
 * @property {function(string, string): {success: boolean, result: boolean, error?: string}} lt
 * @property {function(string, string): {success: boolean, result: boolean, error?: string}} eq
 * @property {function(string, string): {success: boolean, result: boolean, error?: string}} gte
 * @property {function(string, string): {success: boolean, result: boolean, error?: string}} lte
 * @property {function(string, string): {success: boolean, result: boolean, error?: string}} satisfies
 * @property {function(string): {success: boolean, result: string, error?: string}} normalize
 * @property {function(string): {success: boolean, result: string, error?: string}} detectType
 * @property {function(string): SemverConstraint} parseConstraint
 */

/**
 * @typedef {Object} ChecksumResult
 * @property {boolean} success
 * @property {string} [hash]
 * @property {boolean} [valid]
 * @property {string} [actualHash]
 * @property {string} [error]
 */

/**
 * @typedef {Object} Checksum
 * @property {function(string): ChecksumResult} md5
 * @property {function(string): ChecksumResult} sha256
 * @property {function(string): ChecksumResult} sha512
 * @property {function(string, string, string=): ChecksumResult} verify
 * @property {function(string, string=): ChecksumResult} string
 */

/**
 * App 基类
 * @class
 * @global
 * @property {AppMetadata} metadata
 */
class App {
  /**
   * @param {AppMetadata} metadata
   */
  constructor(metadata) {
    /** @type {AppMetadata} */
    this.metadata = metadata;
  }

  /**
   * 检查最新版本
   * @returns {string}
   */
  checkVersion() {
    throw new Error("Not implemented");
  }

  /**
   * 获取下载信息
   * @param {string} version
   * @param {string} arch
   * @returns {DownloadInfo}
   */
  getDownloadInfo(version, arch) {
    throw new Error("Not implemented");
  }

  /**
   * 安全调用钩子函数
   * @template T
   * @param {function(): T} fn
   * @returns {{success: boolean, value?: T, error?: string}}
   */
  safeCall(fn) {
    try {
      const value = fn();
      return { success: true, value };
    } catch (error) {
      return { success: false, error: String(error) };
    }
  }
}

/**
 * 全局 log 对象
 * @type {Log}
 */
let log;

/**
 * 全局 path 对象
 * @type {Pathx}
 */
let path;

/**
 * 全局 exec 对象
 * @type {Execx}
 */
let exec;

/**
 * 全局 fetch 对象
 * @type {Fetch}
 */
let fetch;

/**
 * 全局 fs 对象
 * @type {Fsutil}
 */
let fs;

/**
 * 全局 archive 对象
 * @type {Archive}
 */
let archive;

/**
 * 全局 symlink 对象
 * @type {Symlink}
 */
let symlink;

/**
 * 全局 registry 对象
 * @type {Registry}
 */
let registry;

/**
 * 全局 installer 对象
 * @type {Installer}
 */
let installer;

/**
 * 全局 chopsticks 对象
 * @type {Chopsticks}
 */
let chopsticks;

/**
 * 全局 json 对象
 * @type {Jsonx}
 */
let json;

/**
 * 全局 semver 对象
 * @type {Semver}
 */
let semver;

/**
 * 全局 checksum 对象
 * @type {Checksum}
 */
let checksum;

/**
 * @global
 * @type {typeof App}
 */
let App;
