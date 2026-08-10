export const handlers = {
    executeOperation: async (params: unknown) => {
        console.log('executeOperation', params);
    },

    invalidateQuery: async (params: unknown) => {
        console.log('invalidateQuery', params);
    },
};
