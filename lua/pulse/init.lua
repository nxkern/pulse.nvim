local M = {}
function M.setup(opts)
    opts = require("pulse.strictOpts").setup(opts)

    local bin = require("pulse.binary").setup()
    if not bin or vim.fn.executable(bin) == 0 then
        vim.notify("Pulse: Binary not found or download failed", vim.log.levels.ERROR)
        return
    end
    local server_url = opts.server_url or ""
    local validate = opts.validate or ""
    local isLocalhost = opts.isLocalhost or false
    local state = require("pulse.state")
    local cmd = { bin, server_url, validate, isLocalhost }
    if state.job_id then
        vim.fn.jobstop(state.job_id)
    end

    if state.pulse_idle_timer then
        state.pulse_idle_timer:stop()
    end

    state.job_id = vim.fn.jobstart(cmd, {
        on_stderr = function(_, d, _)
            if d[1] ~= "" then
                vim.notify("Error:" .. d[1], vim.log.levels.ERROR)
            end
        end,
        on_stdout = function(_, d, _)
            vim.notify(d[1], vim.log.levels.INFO)
        end,
        on_exit = function(_, c, _)
            state.job_id = nil
            vim.notify("Process closed with code: " .. c, vim.log.levels.WARN)
        end,
    })

    require("pulse.wsServer.wsServer").setup(opts, state)
    require("pulse.discord.discordIPC").setup(state)
end

return M
