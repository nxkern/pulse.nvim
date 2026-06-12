local M = {};

local state = require("pulse.state");

function M.setup()
    if not state.job_id or state.job_id <= 0 then
        return
    end
    vim.schedule_wrap(function()
        if vim.fn.chansend(state.job_id, '{"status": "testing", "file": ""}') == 1 then
            state.countError = state.countError + 1;
            return
        end
    end)
end

return M
