<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { token, userId, userName } from '../../utils/account';
import { verifyUserInfo } from '../../utils/api/auth';
import { getPbDetails } from '../../utils/api/problems';
import {
    createProblemTestcase,
    deleteProblemTestcase,
    getProblemTestcases,
    reorderProblemTestcases,
    updateProblemTestcase
} from '../../utils/api/testcases';
import { resolveApiErrorMessage } from '../../utils/api/errors';
import { showAlert } from '../../utils/alert';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const loading = ref(true);
const saving = ref(false);
const role = ref('');
const problem = ref({});
const testcases = ref([]);
const editorDialog = ref(false);
const formMode = ref('create');
const form = ref({
    id: '',
    input_data: '',
    output_data: '',
    is_sample: false
});

const problemId = computed(() => Number(route.params.problem_id));
const userLoggedIn = computed(() => !!token.value);
const canManage = computed(() => {
    const currentRole = String(role.value || '').trim();
    const ownerUserID = String(problem.value.owner_user_id || '').trim();
    return currentRole === 'admin' || (currentRole === 'teacher' && ownerUserID === String(userId.value || '').trim());
});

const resetForm = () => {
    form.value = {
        id: '',
        input_data: '',
        output_data: '',
        is_sample: false
    };
};

const loadPage = async () => {
    if (!userLoggedIn.value) {
        window.location = '#/login';
        return;
    }

    loading.value = true;
    try {
        const [verifyResp, problemResp, testcaseResp] = await Promise.all([
            verifyUserInfo(userName.value, token.value),
            getPbDetails(problemId.value),
            getProblemTestcases(problemId.value)
        ]);

        role.value = verifyResp?.data?.data?.role || '';
        problem.value = problemResp?.data?.data || {};
        testcases.value = testcaseResp?.data?.data || [];

        if (!canManage.value) {
            window.location = '#/403';
        }
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'You do not have permission to manage testcases',
            404: 'Problem or testcase API not available'
        }, 'Load testcase management failed'), '/admin/problem');
    } finally {
        loading.value = false;
    }
};

const openCreateDialog = () => {
    formMode.value = 'create';
    resetForm();
    editorDialog.value = true;
};

const openEditDialog = (item) => {
    formMode.value = 'edit';
    form.value = {
        id: item.id,
        input_data: item.input_data || '',
        output_data: item.output_data || '',
        is_sample: Boolean(item.is_sample)
    };
    editorDialog.value = true;
};

const saveTestcase = async () => {
    if (!String(form.value.input_data || '').trim() || !String(form.value.output_data || '').trim()) {
        showAlert('Input and output are required', '');
        return;
    }

    saving.value = true;
    try {
        const payload = {
            input_data: form.value.input_data,
            output_data: form.value.output_data,
            is_sample: Boolean(form.value.is_sample)
        };
        if (formMode.value === 'edit' && form.value.id) {
            await updateProblemTestcase(problemId.value, form.value.id, payload);
        } else {
            await createProblemTestcase(problemId.value, payload);
        }
        editorDialog.value = false;
        resetForm();
        await loadPage();
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'You do not have permission to manage testcases',
            404: 'Problem or testcase not found'
        }, 'Save testcase failed'), '');
    } finally {
        saving.value = false;
    }
};

const removeTestcase = async (item) => {
    saving.value = true;
    try {
        await deleteProblemTestcase(problemId.value, item.id);
        await loadPage();
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'You do not have permission to delete this testcase',
            404: 'Testcase not found'
        }, 'Delete testcase failed'), '');
    } finally {
        saving.value = false;
    }
};

const moveTestcase = async (index, direction) => {
    const target = index + direction;
    if (target < 0 || target >= testcases.value.length) {
        return;
    }
    const next = [...testcases.value];
    const current = next[index];
    next[index] = next[target];
    next[target] = current;
    const orderedIds = next.map((item) => item.id);

    saving.value = true;
    try {
        await reorderProblemTestcases(problemId.value, orderedIds);
        testcases.value = next;
        await loadPage();
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'You do not have permission to reorder testcases',
            404: 'Testcase not found'
        }, 'Reorder testcase failed'), '');
    } finally {
        saving.value = false;
    }
};

onMounted(loadPage);
</script>

<template>
    <div v-if="loading" class="loading">
        <v-progress-circular indeterminate color="primary" :width="12" :size="100"></v-progress-circular>
    </div>
    <div v-else>
        <v-app-bar :elevation="2">
            <template v-slot:prepend>
                <v-btn icon="mdi-chevron-left" size="x-large" @click="router.back()"></v-btn>
            </template>
            <v-col class="align-left">
                <p style="font-size: 24px;">Manage Testcases</p>
                <p>{{ problem.title }}</p>
            </v-col>
            <template v-slot:append>
                <v-btn color="primary" rounded="xl" @click="openCreateDialog">Add Testcase</v-btn>
            </template>
        </v-app-bar>

        <v-container fluid class="pa-6">
            <v-card rounded="xl" elevation="2">
                <template v-slot:title>
                    <span class="font-weight-black">Problem Testcases</span>
                </template>
                <v-card-text>
                    <div v-if="testcases.length === 0" class="empty-text">
                        No testcases configured yet.
                    </div>
                    <v-list v-else>
                        <v-list-item v-for="(item, index) in testcases" :key="item.id">
                            <template v-slot:prepend>
                                <v-chip color="primary" variant="tonal">{{ index + 1 }}</v-chip>
                            </template>
                            <v-list-item-title class="font-weight-medium">
                                {{ item.is_sample ? 'Sample testcase' : 'Judge testcase' }}
                            </v-list-item-title>
                            <v-list-item-subtitle>
                                Input {{ String(item.input_data || '').length }} chars, Output {{ String(item.output_data || '').length }} chars
                            </v-list-item-subtitle>
                            <template v-slot:append>
                                <div class="actions">
                                    <v-btn icon="mdi-arrow-up" variant="text" :disabled="index === 0" @click="moveTestcase(index, -1)"></v-btn>
                                    <v-btn icon="mdi-arrow-down" variant="text" :disabled="index === testcases.length - 1" @click="moveTestcase(index, 1)"></v-btn>
                                    <v-btn icon="mdi-pencil" variant="text" color="primary" @click="openEditDialog(item)"></v-btn>
                                    <v-btn icon="mdi-delete-outline" variant="text" color="error" @click="removeTestcase(item)"></v-btn>
                                </div>
                            </template>
                        </v-list-item>
                    </v-list>
                </v-card-text>
            </v-card>
        </v-container>

        <v-dialog v-model="editorDialog" max-width="900px">
            <v-card rounded="xl">
                <v-card-title>{{ formMode === 'edit' ? t('message.edit') : t('message.addTestCase') }}</v-card-title>
                <v-card-text>
                    <v-switch
                        v-model="form.is_sample"
                        color="primary"
                        inset
                        label="Expose as sample testcase"
                    ></v-switch>
                    <v-row>
                        <v-col cols="12" md="6">
                            <v-textarea
                                v-model="form.input_data"
                                label="Input"
                                rows="12"
                                auto-grow
                                variant="outlined"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-textarea
                                v-model="form.output_data"
                                label="Output"
                                rows="12"
                                auto-grow
                                variant="outlined"
                            />
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer></v-spacer>
                    <v-btn rounded="xl" @click="editorDialog = false">Cancel</v-btn>
                    <v-btn color="primary" rounded="xl" :loading="saving" @click="saveTestcase">Save</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </div>
</template>

<style scoped>
.loading {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
}

.align-left {
    text-align: left;
}

.actions {
    display: flex;
    align-items: center;
    gap: 4px;
}

.empty-text {
    padding: 24px 0;
    text-align: center;
    color: rgba(0, 0, 0, 0.56);
}
</style>
