local M = {}

function M.trim(s)
    return s:match("^%s*(.-)%s*$")
end

function M.split(s, delimiter)
    local result = {}
    for match in string.gmatch(s, "([^" .. delimiter .. "]+)") do
        table.insert(result, match)
    end
    return result
end

function M.starts_with(str, prefix)
    return string.sub(str, 1, string.len(prefix)) == prefix
end

function M.ends_with(str, suffix)
    return suffix == "" or string.sub(str, -string.len(suffix)) == suffix
end

return M
