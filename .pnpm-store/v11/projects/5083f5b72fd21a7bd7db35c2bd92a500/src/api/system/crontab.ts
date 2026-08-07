import request from "@/utils/request";

export interface CrontabInternalTaskOption {
    label: string;
    value: string;
    parameterHint?: string;
}

/**
 * 任务列表
 */
export const crontabListApi = (params?: Record<string, unknown>) => {
    return request.get('/system/crontab/index', { params });
}

/**
 * 创建任务
 */
export const crontabCreateApi = (data: Record<string, unknown>) => {
    return request.post('/system/crontab', data);
}

/**
 * 更新任务
 */
export const crontabUpdateApi = (id: number | string, data: Record<string, unknown>) => {
    return request.put(`/system/crontab/${id}`, data);
}

/**
 * 删除任务
 */
export const crontabDeleteApi = (id: number | string) => {
    return request.delete(`/system/crontab/${id}`);
}

/**
 * 运行一次任务
 */
export const crontabRunApi = (id: number | string) => {
    return request.post(`/system/crontab/run/${id}`);
}

/**
 * 已注册的系统内部任务目录
 */
export const crontabInternalTasksApi = () => {
    return request.get<CrontabInternalTaskOption[]>('/system/crontab/internal-tasks');
}


/**
 * 任务日志列表
 */
export const crontabLogListApi = (params?: Record<string, unknown>) => {
    return request.get('/system/crontab/log/index', { params });
}