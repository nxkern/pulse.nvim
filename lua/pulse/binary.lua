local M = {}

local function get_os_arch()
    local uname = vim.uv.os_uname()
    local os = uname.sysname:lower()
    local arch = uname.machine
    local arch_map = { x86_64 = "amd64", aarch64 = "arm64" }
    arch = arch_map[arch] or arch
    return os, arch
end

local function fetch_latest_tag()
    local ok, resp = pcall(vim.fn.system,
        { "curl", "-s", "https://codeberg.org/api/v1/repos/Nxus/pulse.nvim/releases/latest" })
    if not ok then return "" end
    local ok2, data = pcall(vim.json.decode, resp)
    if ok2 and data and data.tag_name then
        return data.tag_name
    end
    return ""
end

local function download_binary(tag, bin_dir, bin_path, os_name)
    local _, arch = get_os_arch()
    local binary_name = string.format("pulse.nvim_%s_%s", os_name, arch)
    local url = string.format("https://codeberg.org/Nxus/pulse.nvim/releases/download/%s/%s.tar.gz", tag, binary_name)

    vim.fn.system({ "sh", "-c", string.format("curl -sL --fail %s | tar -xzf - -C %s", url, bin_dir) })
    if vim.v.shell_error ~= 0 then return false end

    if os_name ~= "windows_nt" then
        vim.fn.system("chmod +x " .. bin_path)
    end

    return vim.fn.executable(bin_path) == 1
end

local function read_version_file(path)
    local f = io.open(path, "r")
    if not f then return "" end
    local v = f:read("*l")
    f:close()
    return v or ""
end

local function write_version_file(path, tag)
    local f = io.open(path, "w")
    if f then
        f:write(tag)
        f:close()
    end
end

function M.setup()
    local info = debug.getinfo(1, "S").source
    if info:sub(1, 1) ~= "@" then return nil end

    local current_file = info:sub(2)
    local plugin_root = vim.fn.fnamemodify(current_file, ":p:h:h:h")
    local bin_dir = plugin_root .. "/bin"
    local bin_path = bin_dir .. "/pulse"
    local version_file = bin_dir .. "/.version"
    local os_name, _ = get_os_arch()

    vim.fn.mkdir(bin_dir, "p")

    -- fast path: binary exists
    if vim.fn.executable(bin_path) == 1 then
        vim.defer_fn(function()
            local tag = fetch_latest_tag()
            if tag == "" then return end
            local local_ver = read_version_file(version_file)
            if tag ~= local_ver then
                if download_binary(tag, bin_dir, bin_path, os_name) then
                    write_version_file(version_file, tag)
                end
            end
        end, 5000)
        return bin_path
    end

    -- binary doesn't exist: download synchronously
    local tag = fetch_latest_tag()
    if tag ~= "" and download_binary(tag, bin_dir, bin_path, os_name) then
        write_version_file(version_file, tag)
        return bin_path
    end

    vim.notify("Pulse: Binary not found and download failed", vim.log.levels.ERROR)
    return nil
end

return M
