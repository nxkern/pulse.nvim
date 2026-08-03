local M = {}
function M.setup(opts, state)
    opts = require("pulse.strictOpts").setup(opts)

    local commands = require("pulse.commands")

    local path = vim.fn.getcwd()
    local workspace = vim.fn.fnamemodify(path, ":t")

    if state.pulse_idle_timer then
        if not state.pulse_idle_timer:is_closing() then
            state.pulse_idle_timer:stop()
            state.pulse_idle_timer:close()
        end
    end
    local uv = vim.uv or vim.loop
    state.pulse_idle_timer = uv.new_timer()
    state.pulse_timer = uv.new_timer()


    state.pulse_timer:start(
        20000,
        20000,
        -- if state.job_id then
        --     vim.fn.chansend(state.job_id, '{ "status": "", "file": ""}\n')
        -- end
        vim.schedule_wrap(function()
            if state.job_id then
                local success, _ = pcall(vim.fn.chansend, state.job_id, '{ "status": "", "file": "", "id": "wsServer"}\n')
                if success then
                    state.countError = 0
                else
                    state.countError = state.countError + 1
                end
            end
        end)
    )

    if state.job_id then vim.fn.chansend(state.job_id,
            string.format('{ "status": "online", "file": "%s", "id": "wsServer"}\n', workspace)) end
    state.currentStatus = "online"

    if not state.pulse_idle_timer then
        return
    end
    commands.setup(workspace, opts.idle_time or 30000, opts)
end

return M
