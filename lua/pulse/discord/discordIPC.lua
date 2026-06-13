local M = {}
function M.setup(state)
    local path = vim.fn.getcwd()
    local data = vim.json.encode({
        details = "Editando" .. vim.fn.expand("%:t"),
        state = "Project: " .. vim.fn.fnamemodify(path, ":t"),
        large_image = "https://www.iconarchive.com/show/papirus-apps-icons-by-papirus-team/nvim-icon.html"
    })
    if state.job_id then
        local success, _ = pcall(vim.fn.chansend, state.job_id, "connect:" .. data .. "\n")
        if not success then vim.notify("error chansend discord") end
    end
end

return M
