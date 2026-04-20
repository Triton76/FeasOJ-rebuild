export const resolveApiErrorMessage = (error, statusMessages = {}, fallback = 'Request failed') => {
    const status = Number(error?.response?.status || 0);
    if (statusMessages[status]) {
        return statusMessages[status];
    }
    if (status === 400) return 'Invalid request payload';
    if (status === 401) return 'Please login first';
    if (status === 403) return 'Permission denied';
    if (status === 404) return 'Resource not found';
    if (status === 409) return 'Resource state conflict';
    if (status === 429) return 'Too many requests, please retry later';

    const backendMessage = error?.response?.data?.error || error?.response?.data?.message;
    if (backendMessage) {
        return String(backendMessage);
    }
    return fallback;
};
