local M = {
	currentStatus = "offline",
	job_id = nil,
	---@type uv.uv_timer_t|nil
	pulse_idle_timer = nil,
	---@type uv.uv_timer_t|nil
	pulse_timer = nil,
	---@type uv.uv_timer_t|nil
    pulse_timer_discord = nil,
	---@type uv.uv_timer_t|nil
    pulse_idle_timer_discord = nil,
    current_status_discord = "online",
    countError = 0,
}

return M
