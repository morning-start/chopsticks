/**
 * Chopsticks App 类型定义
 * @module @chopsticks/core
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
 * @typedef {Object} Path
 * @property {function(...string): string} join
 * @property {function(string): string} abs
 * @property {function(string): string} dir
 * @property {function(string): string} base
 * @property {function(string): string} ext
 * @property {function(string): boolean} exists
 * @property {function(string): boolean} isDir
 */

/**
 * @typedef {Object} ExecResult
 * @property {number} exitCode
 * @property {string} stdout
 * @property {string} stderr
 * @property {boolean} success
 */

/**
 * @typedef {Object} Exec
 * @property {function(string, ...string): Promise<ExecResult>} exec
 * @property {function(string): Promise<ExecResult>} shell
 * @property {function(string): Promise<ExecResult>} powershell
 */

/**
 * @typedef {Object} FetchResponse
 * @property {number} status
 * @property {boolean} ok
 * @property {string} body
 * @property {Object.<string, string>} headers
 */

/**
 * @typedef {Object} FetchOptions
 * @property {Object.<string, string>} [headers]
 * @property {number} [timeout]
 */

/**
 * @typedef {Object} Fetch
 * @property {function(string, FetchOptions=): Promise<FetchResponse>} get
 * @property {function(string, string=, string=): Promise<FetchResponse>} post
 * @property {function(string, string): Promise<void>} download
 */

/**
 * @typedef {Object} FS
 * @property {function(string, string=): string} readFile
 * @property {function(string, string): void} writeFile
 * @property {function(string, string): void} copy
 * @property {function(string): void} remove
 * @property {function(string): void} removeAll
 * @property {function(string): void} mkdir
 * @property {function(string): void} mkdirAll
 * @property {function(string): string[]} readDir
 * @property {function(string): boolean} exists
 * @property {function(string): boolean} isDir
 * @property {function(string): boolean} isFile
 */

/**
 * @typedef {Object} Archive
 * @property {function(string, string): Promise<void>} extractZip
 * @property {function(string, string): Promise<void>} extract7z
 * @property {function(string, string): Promise<void>} extractTarGz
 * @property {function(string, string): Promise<void>} extract
 */

/**
 * @typedef {Object} Symlink
 * @property {function(string, string): Promise<void>} create
 * @property {function(string, string): Promise<void>} createDir
 * @property {function(string, string): Promise<void>} createHard
 * @property {function(string, string): Promise<void>} createJunction
 * @property {function(string): string} readLink
 * @property {function(string): boolean} isLink
 */

/**
 * @typedef {Object} Registry
 * @property {function(string, string, string): Promise<void>} setValue
 * @property {function(string, string, number): Promise<void>} setDword
 * @property {function(string, string): Promise<string>} getValue
 * @property {function(string, string): Promise<void>} deleteValue
 * @property {function(string): Promise<void>} createKey
 * @property {function(string): Promise<void>} deleteKey
 */

/**
 * @typedef {'nsis'|'msi'|'inno'|'unknown'} InstallerType
 */

/**
 * @typedef {Object} Installer
 * @property {function(string, string[]=): Promise<void>} run
 * @property {function(string, string[]=): Promise<void>} runNSIS
 * @property {function(string, string[]=): Promise<void>} runMSI
 * @property {function(string, string[]=): Promise<void>} runInno
 * @property {function(string): Promise<InstallerType>} detectType
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
 * @typedef {Object} Chopsticks
 * @property {function(string, string): string} getCookDir
 * @property {function(): string} getCacheDir
 * @property {function(): string} getConfigDir
 * @property {function(string, string): Promise<void>} setEnv
 * @property {function(string): Promise<string>} getEnv
 * @property {function(string): Promise<void>} deleteEnv
 * @property {function(string): Promise<void>} addToPath
 * @property {function(string): Promise<void>} removeFromPath
 * @property {function(): string[]} getPath
 * @property {function(ShortcutOptions): Promise<void>} createShortcut
 * @property {function(string, string[]): Promise<void>} persistData
 */

/**
 * @typedef {Object} JSON
 * @property {function(Object): string} stringify
 * @property {function(string): Object} parse
 */

/**
 * @typedef {Object} Semver
 * @property {function(string, string): -1|0|1} compare
 * @property {function(string, string): boolean} gt
 * @property {function(string, string): boolean} lt
 * @property {function(string, string): boolean} eq
 */

/**
 * @typedef {Object} Checksum
 * @property {function(string): Promise<string>} sha256
 * @property {function(string): Promise<string>} md5
 * @property {function(string, string, string): Promise<boolean>} verify
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
     * @returns {Promise<string>}
     */
    async checkVersion() {
        throw new Error('Not implemented');
    }

    /**
     * 获取下载信息
     * @param {string} version
     * @param {string} arch
     * @returns {Promise<DownloadInfo>}
     */
    async getDownloadInfo(version, arch) {
        throw new Error('Not implemented');
    }

    /**
     * 安全调用钩子函数
     * @template T
     * @param {function(): Promise<T>} fn
     * @returns {Promise<{success: boolean, value?: T, error?: string}>}
     */
    async safeCall(fn) {
        try {
            const value = await fn();
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
 * @type {Path}
 */
let path;

/**
 * 全局 exec 对象
 * @type {Exec}
 */
let exec;

/**
 * 全局 fetch 对象
 * @type {Fetch}
 */
let fetch;

/**
 * 全局 fs 对象
 * @type {FS}
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
 * @type {JSON}
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
