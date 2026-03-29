<?php

declare(strict_types=1);

namespace AnyVali\Interchange;

use AnyVali\AnyValiDocument;
use AnyVali\ExportMode;
use AnyVali\Schema;

final class Exporter
{
    private function __construct()
    {
    }

    public static function export(Schema $schema, ExportMode $mode = ExportMode::Portable): AnyValiDocument
    {
        if ($mode === ExportMode::Portable && $schema->hasCustomValidators()) {
            throw new \RuntimeException(
                'Cannot export schema with custom validators in portable mode'
            );
        }

        return new AnyValiDocument(
            root: $schema->exportNode(),
        );
    }
}
