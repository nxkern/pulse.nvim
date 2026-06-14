local M = {}
function M.setup(state)
    local path = vim.fn.getcwd()
    vim.api.nvim_create_autocmd({ 'BufEnter', 'BufReadPost' }, {
        callback = function()
            if vim.bo.filetype == "minifiles" then
                vim.notify("In minifiles", vim.log.levels.WARN)
                return
            end
            local data = vim.json.encode({
                details = (vim.fn.expand("%:t") == "" and "HomePage" or "Editing: " .. vim.fn.expand("%:t")),
                state = "Workspace: " .. vim.fn.fnamemodify(path, ":t"),
                assets = {
                    large_image = string.format(
                        "https://raw.githubusercontent.com/circles-00/nvim-discord-status/refs/heads/main/assets/icons/%s.png",
                        (vim.bo.filetype == "" and "idle" or vim.bo.filetype)),
                    small_image = "https://icons.iconarchive.com/icons/papirus-team/papirus-apps/256/nvim-icon.png",
                }
            })
            if state.job_id then
                local success, _ = pcall(vim.fn.chansend, state.job_id, "connect:" .. data .. "\n")
                if not success then vim.notify("error chansend autocmd discord", vim.log.levels.ERROR) end
            end
        end,
    })
end

return M
