<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { applyJoinClass } from '../../utils/api/classes';
import { showAlert } from '../../utils/alert';
import { resolveApiErrorMessage } from '../../utils/api/errors';

const router = useRouter();
const classCode = ref('');
const loading = ref(false);

const joinClass = async () => {
    if (!classCode.value.trim()) {
        showAlert('Class code is required', '');
        return;
    }

    loading.value = true;
    try {
        await applyJoinClass(classCode.value.trim());
        showAlert('Join request submitted', '/classes');
    } catch (error) {
        const status = Number(error?.response?.status || 0);
        showAlert(resolveApiErrorMessage(error, {
            401: 'Please login first',
            403: 'You do not have permission to join this class',
            404: 'Class code not found',
            409: 'You already have a membership record for this class'
        }, 'Join class failed, please retry'), status === 401 ? '/login' : '');
    } finally {
        loading.value = false;
    }
};
</script>

<template>
    <v-container fluid class="pa-6">
        <v-row justify="center">
            <v-col cols="12" md="8" lg="6">
                <v-card elevation="2" rounded="lg">
                    <v-card-title class="text-h6">Join Class</v-card-title>
                    <v-divider></v-divider>
                    <v-card-text>
                        <v-text-field
                            v-model="classCode"
                            label="Class Code"
                            variant="solo-filled"
                            density="comfortable"
                            prepend-inner-icon="mdi-key-variant"
                        />
                    </v-card-text>
                    <v-card-actions>
                        <v-btn variant="text" @click="router.push('/classes')">Back</v-btn>
                        <v-spacer></v-spacer>
                        <v-btn color="primary" :loading="loading" @click="joinClass">Submit Request</v-btn>
                    </v-card-actions>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>
