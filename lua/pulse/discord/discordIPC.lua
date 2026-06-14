local M = {}
function M.setup(state, opts)
    local path = vim.fn.getcwd()
    local command = require("pulse.discord.commands")
    if state.pulse_idle_timer_discord then
        if not state.pulse_idle_timer_discord:is_closing() then
            state.pulse_idle_timer_discord:stop()
            state.pulse_idle_timer_discord:close()
        end
    end
    local uv = vim.uv or vim.loop
    state.pulse_idle_timer_discord = uv.new_timer()
    state.pulse_timer_discord = uv.new_timer()
    vim.api.nvim_create_autocmd("BufEnter", {
        callback = function()
            local file = vim.fn.expand("%:t")
            if vim.bo.buftype ~= "" and vim.bo.filetype ~= "" then
                file = vim.bo.filetype
            end
            local data = vim.json.encode({
                workspace = vim.fn.fnamemodify(path, ":t"),
                file = file
            })

            if state.job_id then
                local success, _ = pcall(vim.fn.chansend, state.job_id, "connect:" .. data .. "\n")
                if not success then vim.notify("error chansend autocmd discord", vim.log.levels.ERROR) end
            end
        end,
    })
    command.setup(opts.idle_time or 30000)
end

return M
