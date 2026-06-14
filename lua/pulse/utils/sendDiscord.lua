local M = {}

function M.setup(state, file)
    local path = vim.fn.getcwd()
    if state.job_id then
        local data = vim.json.encode({
            workspace = vim.fn.fnamemodify(path, ":t"),
            file = file
        })
        vim.fn.chansend(state.job_id, "connect:" .. data .. "\n")
    end
end

return M
