local M = {}
local state = require("pulse.state")

local function reset_idle_timer(workspace, idle_time)
    if not state.pulse_idle_timer or state.pulse_idle_timer:is_closing() then
        return
    end

    state.pulse_idle_timer:stop()

    if state.currentStatus ~= "online" and state.currentStatus ~= "busy" then
        vim.fn.chansend(state.job_id, string.format('{ "status": "online", "file": "%s"}\n', workspace))
        state.currentStatus = "online"
    end

    state.pulse_idle_timer:start(
        idle_time,
        0,
        vim.schedule_wrap(function()
            if state.currentStatus ~= "idle" and state.currentStatus ~= "busy" then
                vim.fn.chansend(state.job_id, '{ "status": "idle", "file": ""}\n')
                state.currentStatus = "idle"
            end
        end)
    )
end

---@param workspace string
---@param idle_time number
---@param opts table
function M.setup(workspace, idle_time, opts)
    reset_idle_timer(workspace, idle_time)

    vim.api.nvim_create_autocmd({ "TextChanged", "CursorMoved", "CursorMovedI" }, {
        callback = function()
            vim.schedule(function()
                reset_idle_timer(workspace, idle_time)
            end)
        end,
    })

    vim.keymap.set("n", opts.keymaps.idle, function()
        if state.job_id and state.job_id > 0 then
            vim.fn.chansend(state.job_id, '{ "status": "idle", "file": ""}\n')
            state.currentStatus = "idle"
            vim.notify("Pulse: status idle", vim.log.levels.INFO)
        end
    end, { desc = "󰒲 Set status to idle" })

    vim.keymap.set("n", opts.keymaps.busy, function()
        if state.job_id and state.job_id > 0 then
            vim.fn.chansend(state.job_id, '{ "status": "busy", "file": ""}\n')
            state.currentStatus = "busy"
            vim.notify("Pulse: status busy", vim.log.levels.INFO)
        end
    end, { desc = "󰗖 Set status to busy" })

    vim.keymap.set("n", opts.keymaps.online, function()
        if state.job_id and state.job_id > 0 then
            vim.fn.chansend(state.job_id, string.format('{ "status": "online", "file": "%s"}\n', workspace))
            state.currentStatus = "online"
            vim.notify("Pulse: status online", vim.log.levels.INFO)
        end
    end, {
        desc = "󰄬 Set status to online",
    })

    vim.api.nvim_create_autocmd("VimLeavePre", {
        callback = function()
            if state.job_id and state.job_id > 0 then
                vim.fn.chansend(state.job_id, '{ "status": "offline", "file": ""}\n')
            end
        end,
    })
end

return M
