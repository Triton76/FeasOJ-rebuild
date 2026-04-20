<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { verifyUserInfo } from '../../utils/api/auth';
import { token, userName } from '../../utils/account';
import { listClassMemberships, listMyMemberships, reviewMembership } from '../../utils/api/classes';
import { showAlert } from '../../utils/alert';
import { resolveApiErrorMessage } from '../../utils/api/errors';

const route = useRoute();
const router = useRouter();

const classId = computed(() => route.params.class_id || '');
const loading = ref(false);
const acting = ref(false);
const role = ref('');
const myMembershipRole = ref('');
const memberships = ref([]);

const canReview = computed(() => role.value === 'admin' || myMembershipRole.value === 'teacher' || myMembershipRole.value === 'assistant');

const loadMemberships = async () => {
    if (!classId.value) {
        showAlert('Missing class id', '/classes');
        return;
    }

    loading.value = true;
    try {
        const [selfResp, listResp] = await Promise.all([
            listMyMemberships(),
            listClassMemberships(classId.value)
        ]);

        const mine = (selfResp?.data?.data || []).find((item) => item.class_id === classId.value && item.status === 'active');
        myMembershipRole.value = mine?.role_in_class || '';
        memberships.value = listResp?.data?.data || [];
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            401: 'Please login first',
            403: 'You cannot view this class membership list',
            404: 'Class not found'
        }, 'Failed to load memberships'), '/classes');
    } finally {
        loading.value = false;
    }
};

const doReview = async (membershipId, approve) => {
    if (!canReview.value) {
        return;
    }

    acting.value = true;
    try {
        await reviewMembership(membershipId, approve);
        await loadMemberships();
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'You do not have permission to review this request',
            404: 'Membership record not found',
            409: 'Membership already reviewed'
        }, 'Review action failed'), '');
    } finally {
        acting.value = false;
    }
};

onMounted(async () => {
    if (!token.value) {
        showAlert('Please login first', '/login');
        return;
    }

    try {
        const verifyResp = await verifyUserInfo(userName.value, token.value);
        role.value = verifyResp?.data?.data?.role || '';
        await loadMemberships();
    } catch (error) {
        showAlert('Identity verification failed', '/login');
    }
});
</script>

<template>
    <v-container fluid class="pa-6">
        <v-card elevation="2" rounded="lg">
            <v-card-title class="d-flex align-center">
                <span>Class Memberships</span>
                <v-chip class="ml-3" size="small" color="primary" variant="tonal">{{ classId }}</v-chip>
                <v-spacer></v-spacer>
                <v-btn variant="text" @click="router.push('/classes')">Back</v-btn>
            </v-card-title>
            <v-divider></v-divider>
            <v-card-text>
                <v-alert v-if="!canReview" type="info" variant="tonal" class="mb-4">
                    You can view membership status, but review actions are only available to class teacher/assistant or admin.
                </v-alert>

                <v-data-table
                    :headers="[
                        { title: 'Membership ID', value: 'id' },
                        { title: 'User ID', value: 'user_id' },
                        { title: 'Role', value: 'role_in_class' },
                        { title: 'Status', value: 'status' },
                        { title: 'Action', value: 'actions', sortable: false }
                    ]"
                    :items="memberships"
                    :loading="loading"
                    item-key="id"
                >
                    <template #item.actions="{ item }">
                        <div class="d-flex ga-2">
                            <v-btn
                                size="small"
                                color="success"
                                variant="tonal"
                                :disabled="!canReview || item.status !== 'pending' || acting"
                                @click="doReview(item.id, true)"
                            >Approve</v-btn>
                            <v-btn
                                size="small"
                                color="error"
                                variant="tonal"
                                :disabled="!canReview || item.status !== 'pending' || acting"
                                @click="doReview(item.id, false)"
                            >Reject</v-btn>
                        </div>
                    </template>
                </v-data-table>
            </v-card-text>
        </v-card>
    </v-container>
</template>
