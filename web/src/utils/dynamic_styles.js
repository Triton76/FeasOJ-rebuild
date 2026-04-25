const RESULT_ALIAS = {
        compile_failed: 'compile_error',
        compilefailed: 'compile_error',
        compile_error: 'compile_error',
        time_limit_exceeded: 'time_limit',
        tle: 'time_limit',
        memory_limit_exceeded: 'memory_limit',
        mle: 'memory_limit',
        wrong_answer: 'wrong_answer',
        wa: 'wrong_answer',
        runtime_error: 'runtime_error',
        re: 'runtime_error',
        internal_error: 'system_error',
        failed: 'system_error',
};

const RESULT_LABEL = {
        pending: 'Pending',
        judging: 'Judging',
        accepted: 'Accepted',
        wrong_answer: 'Wrong Answer',
        time_limit: 'Time Limit Exceeded',
        memory_limit: 'Memory Limit Exceeded',
        runtime_error: 'Runtime Error',
        compile_error: 'Compile Error',
        system_error: 'System Error',
};

export const normalizeSubmissionResult = (result) => {
        const raw = String(result || '').trim().toLowerCase().replace(/\s+/g, '_');
        if (!raw) {
                return '';
        }
        return RESULT_ALIAS[raw] || raw;
};

export const isSubmissionInFlight = (result) => {
        const normalized = normalizeSubmissionResult(result);
        return normalized === 'pending' || normalized === 'judging';
};

export const formatSubmissionResultLabel = (result) => {
        const normalized = normalizeSubmissionResult(result);
        return RESULT_LABEL[normalized] || String(result || 'Unknown');
};

// 根据结果不同显示不同颜色
export const getResultStyle = (result) => {
        const normalized = normalizeSubmissionResult(result);
        switch (normalized) {
                case 'accepted':
                        return 'color: green; font-weight: bold;';
                case 'pending':
                case 'judging':
                        return 'color: #1976d2; font-weight: bold;';
                case 'wrong_answer':
                case 'runtime_error':
                        return 'color: orange; font-weight: bold;';
                case 'time_limit':
                case 'memory_limit':
                case 'compile_error':
                case 'system_error':
                        return 'color: red; font-weight: bold;';
                default:
                        return '';
        }
};

// 根据结果不同显示不同Chip颜色
export const getResultChipColor = (result) => {
    const normalized = normalizeSubmissionResult(result);
    switch (normalized) {
        case 'accepted':
            return 'success';
        case 'pending':
        case 'judging':
            return 'info';
        case 'time_limit':
        case 'memory_limit':
            return 'warning';
        case 'wrong_answer':
        case 'compile_error':
        case 'runtime_error':
        case 'system_error':
            return 'error';
        default:
            return 'default';
    }
}

// 根据题目难度显示不同字体
export const difficultyColor = (difficulty) => {
    switch (difficulty) {
        case 0:
            return 'font-weight: bold;color: green;';
        case 1:
            return 'font-weight: bold;color: orange;';
        case 2:
            return 'font-weight: bold;color: red;';
        default:
            return 'font-weight: bold;color: green;';
    }
};

export const difficultyLang = (difficulty) => {
    switch (difficulty) {
        case 0:
            return 'message.easy';
        case 1:
            return 'message.medium';
        case 2:
            return 'message.hard';
        default:
            return 'message.easy';
    }
}