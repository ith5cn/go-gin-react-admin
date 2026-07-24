import request from "@/utils/request";

export const operLogListApi = (params?: Record<string, unknown>) => request.get("/system/oper-log/index", { params });
export const operLogDeleteApi = (id: string | number) => request.delete(`/system/oper-log/${id}`);