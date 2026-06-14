local M = {}
local state = require("pulse.state")
local util = require("pulse.utils.sendDiscord")

local function reset_idle_timer(idle_time)
    if not state.pulse_idle_timer_discord or state.pulse_idle_timer_discord:is_closing() then
        vim.notify("discord timer guard", vim.log.levels.WARN)
        return
    end
    state.pulse_idle_timer_discord:stop()
    if state.current_status_discord ~= "online" then
        local file = vim.fn.expand("%:t")
        if vim.bo.buftype ~= "" and vim.bo.filetype ~= "" then
            file = vim.bo.filetype
        end
        util.setup(state, file)
        state.current_status_discord = "online"
    end

    state.pulse_idle_timer_discord:start(idle_time, 0, vim.schedule_wrap(function()
        if state.current_status_discord ~= "idle" then
            util.setup(state, "💤")
            state.current_status_discord = "idle"
        end
    end))
end

function M.setup(idle_time)
    reset_idle_timer(idle_time)
    vim.api.nvim_create_autocmd({ "TextChanged", "CursorMoved", "CursorMovedI" }, {
        callback = function()
            vim.schedule(function()
                reset_idle_timer(idle_time)
            end)
        end
    })
end

return M
