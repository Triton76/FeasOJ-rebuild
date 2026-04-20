<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { verifyUserInfo } from '../../utils/api/auth';
import { token, userName } from '../../utils/account';
import { showAlert } from '../../utils/alert';
import { resolveApiErrorMessage } from '../../utils/api/errors';
import {
    listMyMemberships,
    createClass,
    updateClass,
    archiveClass
} from '../../utils/api/classes';

const router = useRouter();

const loading = ref(false);
const saving = ref(false);
const archiving = ref(false);
const role = ref('');
const memberships = ref([]);

const createDialog = ref(false);
const editDialog = ref(false);
const archiveDialog = ref(false);

const createForm = ref({ name: '', code: '', description: '' });
const editForm = ref({ class_id: '', name: '', description: '' });
const archiveClassId = ref('');

const canManageClass = computed(() => role.value === 'teacher' || role.value === 'admin');
const myManagedClassIds = computed(() => (memberships.value || [])
    .filter((item) => item.status === 'active' && (item.role_in_class === 'teacher' || item.role_in_class === 'assistant'))
    .map((item) => item.class_id));

const loadMyMemberships = async () => {
    loading.value = true;
    try {
        const resp = await listMyMemberships();
        memberships.value = resp?.data?.data || [];
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            401: 'Please login first'
        }, 'Failed to load your class memberships'), error?.response?.status === 401 ? '/login' : '');
    } finally {
        loading.value = false;
    }
};

const submitCreate = async () => {
    if (!createForm.value.name.trim() || !createForm.value.code.trim()) {
        showAlert('Class name and code are required', '');
        return;
    }

    saving.value = true;
    try {
        await createClass({
            name: createForm.value.name.trim(),
            code: createForm.value.code.trim(),
            description: createForm.value.description.trim()
        });
        createDialog.value = false;
        createForm.value = { name: '', code: '', description: '' };
        await loadMyMemberships();
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'Only teacher/admin can create classes',
            409: 'Class code already exists'
        }, 'Create class failed'), '');
    } finally {
        saving.value = false;
    }
};

const openEditDialog = (classId) => {
    editForm.value = {
        class_id: classId,
        name: '',
        description: ''
    };
    editDialog.value = true;
};

const submitEdit = async () => {
    if (!editForm.value.class_id.trim()) {
        showAlert('Class id is required', '');
        return;
    }

    saving.value = true;
    try {
        await updateClass(editForm.value.class_id.trim(), {
            name: editForm.value.name.trim(),
            description: editForm.value.description.trim()
        });
        editDialog.value = false;
        await loadMyMemberships();
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'You do not have permission to edit this class',
            404: 'Class not found'
        }, 'Update class failed'), '');
    } finally {
        saving.value = false;
    }
};

const submitArchive = async () => {
    if (!archiveClassId.value.trim()) {
        showAlert('Class id is required', '');
        return;
    }

    archiving.value = true;
    try {
        await archiveClass(archiveClassId.value.trim());
        archiveDialog.value = false;
        archiveClassId.value = '';
        await loadMyMemberships();
    } catch (error) {
        showAlert(resolveApiErrorMessage(error, {
            403: 'Only class owner can archive this class',
            404: 'Class not found'
        }, 'Archive class failed'), '');
    } finally {
        archiving.value = false;
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
        await loadMyMemberships();
    } catch (error) {
        showAlert('Identity verification failed', '/login');
    }
});
</script>

<template>
    <v-container fluid class="pa-6">
        <v-row>
            <v-col cols="12">
                <v-card elevation="2" rounded="lg" class="mb-6">
                    <v-card-title class="d-flex align-center">
                        <span>Class Workflows</span>
                        <v-spacer></v-spacer>
                        <v-btn color="primary" variant="tonal" class="mr-2" @click="router.push('/classes/join')">Join Class</v-btn>
                        <v-btn v-if="canManageClass" color="primary" @click="createDialog = true">Create Class</v-btn>
                    </v-card-title>
                    <v-divider></v-divider>
                    <v-card-text>
                        <v-alert type="info" variant="tonal">
                            This page uses canonical backend class membership APIs. Teacher/assistant/admin actions are shown by role.
                        </v-alert>
                    </v-card-text>
                </v-card>
            </v-col>

            <v-col cols="12">
                <v-card elevation="2" rounded="lg">
                    <v-card-title>My Memberships</v-card-title>
                    <v-divider></v-divider>
                    <v-card-text>
                        <v-data-table
                            :headers="[
                                { title: 'Class ID', value: 'class_id' },
                                { title: 'Role In Class', value: 'role_in_class' },
                                { title: 'Status', value: 'status' },
                                { title: 'Actions', value: 'actions', sortable: false }
                            ]"
                            :items="memberships"
                            :loading="loading"
                            item-key="id"
                        >
                            <template #item.actions="{ item }">
                                <div class="d-flex ga-2 flex-wrap">
                                    <v-btn size="small" variant="tonal" color="primary" @click="router.push(`/classes/${item.class_id}/memberships`)">
                                        Members
                                    </v-btn>
                                    <v-btn
                                        v-if="canManageClass && (item.role_in_class === 'teacher' || item.role_in_class === 'assistant') && item.status === 'active'"
                                        size="small"
                                        variant="tonal"
                                        color="warning"
                                        @click="openEditDialog(item.class_id)"
                                    >
                                        Edit
                                    </v-btn>
                                    <v-btn
                                        v-if="canManageClass && item.role_in_class === 'teacher' && item.status === 'active'"
                                        size="small"
                                        variant="tonal"
                                        color="error"
                                        @click="archiveClassId = item.class_id; archiveDialog = true"
                                    >
                                        Archive
                                    </v-btn>
                                </div>
                            </template>
                        </v-data-table>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>

        <v-dialog v-model="createDialog" max-width="640">
            <v-card>
                <v-card-title>Create Class</v-card-title>
                <v-divider></v-divider>
                <v-card-text>
                    <v-text-field v-model="createForm.name" label="Class Name" variant="solo-filled" class="mb-3" />
                    <v-text-field v-model="createForm.code" label="Class Code" variant="solo-filled" class="mb-3" />
                    <v-textarea v-model="createForm.description" label="Description" variant="solo-filled" />
                </v-card-text>
                <v-card-actions>
                    <v-btn variant="text" @click="createDialog = false">Cancel</v-btn>
                    <v-spacer></v-spacer>
                    <v-btn color="primary" :loading="saving" @click="submitCreate">Create</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>

        <v-dialog v-model="editDialog" max-width="640">
            <v-card>
                <v-card-title>Edit Class</v-card-title>
                <v-divider></v-divider>
                <v-card-text>
                    <v-text-field v-model="editForm.class_id" label="Class ID" variant="solo-filled" readonly class="mb-3" />
                    <v-text-field v-model="editForm.name" label="New Name (optional)" variant="solo-filled" class="mb-3" />
                    <v-textarea v-model="editForm.description" label="New Description (optional)" variant="solo-filled" />
                </v-card-text>
                <v-card-actions>
                    <v-btn variant="text" @click="editDialog = false">Cancel</v-btn>
                    <v-spacer></v-spacer>
                    <v-btn color="primary" :loading="saving" @click="submitEdit">Save</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>

        <v-dialog v-model="archiveDialog" max-width="520">
            <v-card>
                <v-card-title>Archive Class</v-card-title>
                <v-divider></v-divider>
                <v-card-text>
                    <div class="text-body-1 mb-3">This operation archives the class and is only available to the class owner.</div>
                    <v-text-field v-model="archiveClassId" label="Class ID" variant="solo-filled" readonly />
                </v-card-text>
                <v-card-actions>
                    <v-btn variant="text" @click="archiveDialog = false">Cancel</v-btn>
                    <v-spacer></v-spacer>
                    <v-btn color="error" :loading="archiving" @click="submitArchive">Archive</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-container>
</template>
