local M = {};

local function validateKeymap(keymap)
    local OK, result = pcall(vim.api.nvim_replace_termcodes, keymap, true, true, true)
    if not OK then
        vim.notify(string.format("Pulse: '%s' is invalidate format please use this format: <leader>p", keymap))
        return false
    end
    if keymap:match("<") and result == keymap then
        vim.notify(string.format("Pulse: '%s' not recognized", keymap))
        return false
    end
    return true
end

function M.setup(opts)
    opts = opts or {};

    local server = opts.server_url;
    local validate = opts.validate;
    local localHost = opts.isLocalhost;
    local idleTimer = opts.idle_time;
    local keymaps = opts.keymaps;

    if server and type(server) ~= "string" then
        vim.notify("Pulse: 'server_url' must be a string", vim.log.levels.WARN);
        opts.server_url = nil
    end

    if validate and type(validate) ~= "string" then
        vim.notify("Pulse: 'validate' must be a string", vim.log.levels.WARN)
        opts.validate = nil
    end

    if localHost ~= nil and type(localHost) ~= "boolean" then
        vim.notify("Pulse: 'isLocalhost' must be a boolean", vim.log.levels.WARN)
        opts.isLocalhost = nil
    end

    if idleTimer ~= nil and type(idleTimer) ~= "number" then
        vim.notify("Pulse: 'idle_time' must be a number (miliseconds)", vim.log.levels.WARN)
        opts.idle_time = nil
    end

    if keymaps and type(keymaps) ~= "table" then
        vim.notify("Pulse: 'keymaps' must be a table")
        opts.keymaps = nil
    end

    if keymaps.idle ~= nil and type(keymaps.idle) ~= "string" then
        local idle = validateKeymap(keymaps.idle)
        if not idle then
            opts.keymaps.idle = "<leader>Cy"
        end
   end

    if keymaps.busy ~= nil and type(keymaps.busy) ~= "string" then
        local busy = validateKeymap(keymaps.busy)
        if not busy then
            opts.keymaps.busy = "<leader>Cb"
        end
    end

    if keymaps.online ~= nil and type(keymaps.online) ~= "string" then
        local online = validateKeymap(keymaps.online)
        if not online then
            opts.keymaps.online = "<leader>Co"
        end
    end

    return opts
end

return M
