<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { token, userId, userName } from '../../utils/account';
import { verifyUserInfo } from '../../utils/api/auth';
import { getCompetitionById, getCompetitionProblems, replaceCompetitionProblems } from '../../utils/api/competitions';
import { getAllProblemsAdmin } from '../../utils/api/admin';
import { showAlert } from '../../utils/alert';
import { resolveApiErrorMessage } from '../../utils/api/errors';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const loading = ref(true);
const saving = ref(false);
const contestInfo = ref({});
const role = ref('');
const availableProblems = ref([]);
const selectedProblems = ref([]);

const competitionId = computed(() => Number(route.params.competition_id));
const userLoggedIn = computed(() => !!token.value);
const canManage = computed(() => {
    const currentRole = String(role.value || '').trim();
    const ownerUserID = String(contestInfo.value.owner_user_id || '').trim();
    return currentRole === 'admin' || ((currentRole === 'teacher' || currentRole === 'admin') && ownerUserID === String(userId.value || '').trim());
});

const selectedProblemIds = computed(() => new Set(selectedProblems.value.map((item) => Number(item.problem_id))));
const candidateProblems = computed(() => {
    return availableProblems.value.filter((problem) => !selectedProblemIds.value.has(Number(problem.id)));
});

const loadPage = async () => {
    if (!userLoggedIn.value) {
        window.location = '#/login';
        return;
    }

    loading.value = true;
    try {
        const [verifyResp, contestResp, bindingsResp, problemsResp] = await Promise.all([
            verifyUserInfo(userName.value, token.value),
            getCompetitionById(competitionId.value),
            getCompetitionProblems(competitionId.value),
            getAllProblemsAdmin()
        ]);

        role.value = verifyResp?.data?.data?.role || '';
        contestInfo.value = contestResp?.data?.data || {};
        availableProblems.value = problemsResp?.data?.data || [];
        selectedProblems.value = (bindingsResp?.data?.data || []).map((item, index) => ({
            problem_id: Number(item.problem_id),
            alias: item.alias || '',
            display_order: Number(item.display_order || index + 1),
            title: item.title || `Problem ${item.problem_id}`
        }));

        if (!canManage.value) {
            window.location = '#/403';
        }
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            401: 'Please login first',
            403: 'You do not have permission to manage contest problems',
            404: 'Contest not found'
        }, 'Load contest problem management failed'), '/competitions');
    } finally {
        loading.value = false;
    }
};

const addProblem = (problem) => {
    selectedProblems.value.push({
        problem_id: Number(problem.id),
        alias: '',
        display_order: selectedProblems.value.length + 1,
        title: problem.title || `Problem ${problem.id}`
    });
};

const removeProblem = (index) => {
    selectedProblems.value.splice(index, 1);
    selectedProblems.value = selectedProblems.value.map((item, idx) => ({
        ...item,
        display_order: idx + 1
    }));
};

const moveProblem = (index, direction) => {
    const target = index + direction;
    if (target < 0 || target >= selectedProblems.value.length) {
        return;
    }
    const next = [...selectedProblems.value];
    const current = next[index];
    next[index] = next[target];
    next[target] = current;
    selectedProblems.value = next.map((item, idx) => ({
        ...item,
        display_order: idx + 1
    }));
};

const saveBindings = async () => {
    saving.value = true;
    try {
        const items = selectedProblems.value.map((item, index) => ({
            problem_id: Number(item.problem_id),
            alias: String(item.alias || '').trim(),
            display_order: index + 1
        }));
        await replaceCompetitionProblems(competitionId.value, items);
        showAlert('Contest problems saved', `/competitions/${competitionId.value}`);
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'You do not have permission to manage contest problems',
            404: 'Contest or problem not found',
            409: 'Contest problem bindings contain duplicate problem, alias, or order'
        }, 'Save contest problems failed'), '');
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
                <p style="font-size: 24px;">Manage Contest Problems</p>
                <p>{{ contestInfo.title }}</p>
            </v-col>
            <template v-slot:append>
                <v-btn color="primary" rounded="xl" :loading="saving" @click="saveBindings">Save</v-btn>
            </template>
        </v-app-bar>

        <v-container fluid class="pa-6">
            <v-row>
                <v-col cols="12" lg="7">
                    <v-card rounded="xl" elevation="2">
                        <template v-slot:title>
                            <span class="font-weight-black">Selected Problems</span>
                        </template>
                        <v-card-text>
                            <div v-if="selectedProblems.length === 0" class="empty-text">
                                No problems bound yet.
                            </div>
                            <v-list v-else>
                                <v-list-item v-for="(item, index) in selectedProblems" :key="`${item.problem_id}-${index}`">
                                    <template v-slot:prepend>
                                        <v-chip color="primary" variant="tonal">{{ index + 1 }}</v-chip>
                                    </template>
                                    <v-list-item-title>{{ item.title }}</v-list-item-title>
                                    <v-list-item-subtitle>Problem ID: {{ item.problem_id }}</v-list-item-subtitle>
                                    <template v-slot:append>
                                        <div class="selected-actions">
                                            <v-text-field
                                                v-model="item.alias"
                                                label="Alias"
                                                density="compact"
                                                variant="solo-filled"
                                                hide-details
                                                class="alias-field"
                                            />
                                            <v-btn icon="mdi-arrow-up" variant="text" :disabled="index === 0" @click="moveProblem(index, -1)"></v-btn>
                                            <v-btn icon="mdi-arrow-down" variant="text" :disabled="index === selectedProblems.length - 1" @click="moveProblem(index, 1)"></v-btn>
                                            <v-btn icon="mdi-delete-outline" variant="text" color="error" @click="removeProblem(index)"></v-btn>
                                        </div>
                                    </template>
                                </v-list-item>
                            </v-list>
                        </v-card-text>
                    </v-card>
                </v-col>

                <v-col cols="12" lg="5">
                    <v-card rounded="xl" elevation="2">
                        <template v-slot:title>
                            <span class="font-weight-black">Available Problems</span>
                        </template>
                        <v-card-text>
                            <div v-if="candidateProblems.length === 0" class="empty-text">
                                No additional problems available.
                            </div>
                            <v-list v-else>
                                <v-list-item v-for="problem in candidateProblems" :key="problem.id">
                                    <v-list-item-title>{{ problem.title }}</v-list-item-title>
                                    <v-list-item-subtitle>Problem ID: {{ problem.id }}</v-list-item-subtitle>
                                    <template v-slot:append>
                                        <v-btn color="primary" variant="text" @click="addProblem(problem)">Add</v-btn>
                                    </template>
                                </v-list-item>
                            </v-list>
                        </v-card-text>
                    </v-card>
                </v-col>
            </v-row>
        </v-container>
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

.selected-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    min-width: 360px;
}

.alias-field {
    max-width: 120px;
}

.empty-text {
    padding: 24px 0;
    text-align: center;
    color: rgba(0, 0, 0, 0.56);
}
</style>
