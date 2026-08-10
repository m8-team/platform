'use client';

import {defineRegistry} from '@json-render/react';
import {
    Button,
    Card,
    Text,
} from '@gravity-ui/uikit';

import {catalog} from '@/ui/catalog/catalog';

const gaps = {
    xs: 4,
    s: 8,
    m: 16,
    l: 24,
    xl: 32,
} as const;

export const {registry} = defineRegistry(catalog, {
    components: {
        Page: ({props, children}) => {
            const maxWidth =
                props.width === 'normal'
                    ? 1200
                    : props.width === 'wide'
                        ? 1600
                        : undefined;

            return (
                <main
                    style={{
                        width: '100%',
                        maxWidth,
                        margin: '0 auto',
                        padding: 24,
                    }}
                >
                    {children}
                </main>
            );
        },

        Stack: ({props, children}) => (
            <div
                style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: gaps[props.gap],
                }}
            >
                {children}
            </div>
        ),

        Heading: ({props}) => (
            <Text
                as={`h${props.level}`}
                variant={
                    props.level === '1'
                        ? 'display-1'
                        : props.level === '2'
                            ? 'header-2'
                            : 'subheader-3'
                }
            >
                {props.text}
            </Text>
        ),

        Text: ({props}) => (
            <Text
                color={
                    props.tone === 'primary'
                        ? 'primary'
                        : props.tone === 'secondary'
                            ? 'secondary'
                            : props.tone
                }
            >
                {props.text}
            </Text>
        ),

        Card: ({props, children}) => (
            <Card
                type="container"
                view="outlined"
                size="l"
                style={{padding: 20}}
            >
                {props.title ? (
                    <Text
                        as="h3"
                        variant="subheader-3"
                        style={{display: 'block', marginBottom: 16}}
                    >
                        {props.title}
                    </Text>
                ) : null}

                {children}
            </Card>
        ),

        Button: ({props, emit}) => (
            <Button
                view={props.view}
                size="l"
                onClick={() => emit('press')}
            >
                {props.label}
            </Button>
        ),
    },
});
