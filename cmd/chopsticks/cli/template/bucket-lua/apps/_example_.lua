-- Example App
--
-- A sample app

local M = {}

M.metadata = {
    name = "example",
    description = "Example Application",
    homepage = "https://example.com",
    license = "MIT",
    bucket = "{{.Name}}",
}

function M:new()
    local obj = setmetatable({}, { __index = self })
    obj.metadata = self.metadata
    return obj
end

function M:checkVersion()
    return "1.0.0"
end

function M:getDownloadInfo(version, arch)
    return {
        url = string.format("https://example.com/download/%s/app-%s.zip", version, arch),
        type = "zip",
    }
end

function M:onPostInstall(ctx)
    log:info("Installation complete")
end

return
 M
