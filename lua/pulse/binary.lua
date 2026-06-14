local M = {}

function M.setup()
    local info = debug.getinfo(1, "S").source
    if info:sub(1, 1) ~= "@" then
        return nil
    end
    local current_file = info:sub(2)
    local plugin_root = vim.fn.fnamemodify(current_file, ":p:h:h:h")
    local bin_dir = plugin_root .. "/bin"
    local bin_path = bin_dir .. "/pulse"

    vim.fn.mkdir(bin_dir, "p")

    local uname = vim.uv.os_uname()
    local os = uname.sysname:lower()
    local arch = uname.machine
    local arch_map = { x86_64 = "amd64", aarch64 = "arm64" }
    arch = arch_map[arch] or arch
    local binary_name = string.format("pulse.nvim_%s_%s", os, arch)

    local tag_name = ""
    local ok, resp = pcall(vim.fn.system,
        { "curl", "-s", "https://codeberg.org/api/v1/repos/Nxus/pulse.nvim/releases/latest" })
    if ok then
        local ok2, data = pcall(vim.json.decode, resp)
        if ok2 and data and data.tag_name then
            tag_name = data.tag_name
        end
    end

    if tag_name ~= "" then
        local url = string.format("https://codeberg.org/Nxus/pulse.nvim/releases/download/%s/%s.tar.gz", tag_name,
            binary_name)
        local download_cmd = string.format("curl -sL --fail %s | tar -xzf - -C %s", url, bin_dir)
        vim.fn.system({ "sh", "-c", download_cmd })
        if vim.v.shell_error == 0 then
            if os ~= "windows_nt" then
                vim.fn.system("chmod +x " .. bin_path)
            end
            if vim.fn.executable(bin_path) == 1 then
                return bin_path
            end
        end
    end

    if vim.fn.executable(bin_path) == 1 then
        return bin_path
    end

    vim.notify("Pulse: Binary not found and download failed", vim.log.levels.ERROR)
    return nil
end

return M
